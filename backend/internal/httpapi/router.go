package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/activityimport"
	"github.com/Saulorangel87/App-de-treino/backend/internal/ai"
	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"github.com/Saulorangel87/App-de-treino/backend/internal/email"
	"github.com/Saulorangel87/App-de-treino/backend/internal/evolution"
	"github.com/Saulorangel87/App-de-treino/backend/internal/feedback"
	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
)

type Pinger interface{ Ping(context.Context) error }

// SchemaChecker is optionally implemented by the Pinger. When it is, /ready also
// fails while the database is missing migrations the application requires, which
// a plain connectivity check cannot detect.
type SchemaChecker interface{ Check(context.Context) error }

func NewRouter(db Pinger, authService *auth.Service, athleteService *athlete.Service, onboardingService *athlete.OnboardingService, assessmentService *athlete.AssessmentService, recoveryService *athlete.RecoveryService, evolutionService *evolution.Service, feedbackService *feedback.Service, planningService *planning.Service, activityImportService *activityimport.Service, aiService *ai.Service, emailSender email.Sender, appBaseURL, allowedOrigin string, secureCookies, development bool, sessionTTL, emailTokenTTL time.Duration, options ...RouterOption) http.Handler {
	mux := http.NewServeMux()
	server := &Server{auth: authService, athlete: athleteService, onboarding: onboardingService, assessments: assessmentService, recovery: recoveryService, evolution: evolutionService, feedback: feedbackService, planning: planningService, activityImport: activityImportService, ai: aiService, emailSender: emailSender, appBaseURL: appBaseURL, secureCookies: secureCookies, development: development, sessionTTL: sessionTTL, emailTokenTTL: emailTokenTTL}
	server.loginFailures = newRequestRateLimiter()
	for _, option := range options {
		option(server)
	}
	authLimiter := newRequestRateLimiter()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "cadencia-api"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		if checker, ok := db.(SchemaChecker); ok {
			if err := checker.Check(ctx); err != nil {
				slog.Default().Error("readiness schema check failed", "error", err)
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "schema_behind"})
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("POST /v1/auth/register", authLimiter.limit("register", 5, time.Hour, server.register))
	mux.HandleFunc("POST /v1/auth/login", authLimiter.limit("login", 10, 15*time.Minute, server.login))
	mux.HandleFunc("POST /v1/auth/logout", server.logout)
	mux.HandleFunc("DELETE /v1/auth/account", server.deleteAccount)
	mux.HandleFunc("GET /v1/auth/account/export", authLimiter.limit("account-export", 10, time.Hour, server.exportAccountData))
	mux.HandleFunc("POST /v1/auth/accept-terms", server.acceptTerms)
	mux.HandleFunc("POST /v1/auth/change-password", authLimiter.limit("change-password", 5, 15*time.Minute, server.changePassword))
	mux.HandleFunc("POST /v1/auth/logout-others", server.logoutOthers)
	mux.HandleFunc("POST /v1/auth/resend-verification", authLimiter.limit("resend-verification", 5, time.Hour, server.resendVerification))
	mux.HandleFunc("POST /v1/auth/verify-email", authLimiter.limit("verify-email", 20, 15*time.Minute, server.verifyEmail))
	mux.HandleFunc("POST /v1/auth/forgot-password", authLimiter.limit("forgot-password", 5, time.Hour, server.forgotPassword))
	mux.HandleFunc("POST /v1/auth/reset-password", authLimiter.limit("reset-password", 5, 15*time.Minute, server.resetPassword))
	mux.HandleFunc("GET /v1/me", server.me)
	mux.HandleFunc("GET /v1/profile", server.getProfile)
	mux.HandleFunc("PUT /v1/profile", server.putProfile)
	mux.HandleFunc("GET /v1/onboarding", server.getOnboarding)
	mux.HandleFunc("GET /v1/onboarding/questionnaire", server.getQuestionnaire)
	mux.HandleFunc("PUT /v1/onboarding/limitations", server.putLimitations)
	mux.HandleFunc("PUT /v1/onboarding/goals", server.putGoals)
	mux.HandleFunc("PUT /v1/onboarding/availability", server.putAvailability)
	mux.HandleFunc("PUT /v1/onboarding/cycling-context", server.putCyclingContext)
	mux.HandleFunc("GET /v1/assessments/current", server.currentAssessment)
	mux.HandleFunc("POST /v1/assessments/submaximal", server.saveSubmaxAssessment)
	mux.HandleFunc("GET /v1/recovery/today", server.todayRecovery)
	mux.HandleFunc("PUT /v1/recovery/today", server.putTodayRecovery)
	mux.HandleFunc("GET /v1/evolution/summary", server.evolutionSummary)
	mux.HandleFunc("POST /v1/feedback", server.createUserFeedback)
	mux.HandleFunc("POST /v1/workouts/{workoutID}/explanation", server.explainWorkout)
	mux.HandleFunc("POST /v1/plans/generate", server.generatePlan)
	mux.HandleFunc("GET /v1/plans/current", server.currentPlan)
	mux.HandleFunc("GET /v1/activities", server.activities)
	mux.HandleFunc("POST /v1/plans/{planID}/activate", server.activatePlan)
	mux.HandleFunc("POST /v1/protection/recovered", server.reportRecovered)
	mux.HandleFunc("POST /v1/workouts/{workoutID}/start", server.startWorkout)
	mux.HandleFunc("POST /v1/workouts/{workoutID}/complete", server.completeWorkout)
	mux.HandleFunc("POST /v1/workouts/{workoutID}/log", server.logWorkout)
	mux.HandleFunc("POST /v1/workouts/{workoutID}/correct", server.correctWorkout)
	mux.HandleFunc("POST /v1/workouts/{workoutID}/cancel", server.cancelWorkout)
	mux.HandleFunc("POST /v1/workouts/{workoutID}/undo", server.undoWorkout)
	mux.HandleFunc("POST /v1/workouts/{workoutID}/missed", server.markWorkoutMissed)
	mux.HandleFunc("POST /v1/activities/import", server.importActivity)
	mux.HandleFunc("GET /v1/activities/imported", server.listImportedActivities)
	mux.HandleFunc("DELETE /v1/activities/imported/{activityID}", server.deleteImportedActivity)
	mux.HandleFunc("PUT /v1/activities/imported/{activityID}/workout", server.linkImportedActivity)
	mux.HandleFunc("GET /v1/activities/imported/{activityID}/candidates", server.importedActivityCandidates)
	return securityHeaders(secureCookies, observability(slog.Default(), cors(allowedOrigin, csrfProtection(allowedOrigin, mux))))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func cors(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Vary", "Origin")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
