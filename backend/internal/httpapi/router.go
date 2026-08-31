package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
)

type Pinger interface{ Ping(context.Context) error }

func NewRouter(db Pinger, authService *auth.Service, athleteService *athlete.Service, onboardingService *athlete.OnboardingService, planningService *planning.Service, allowedOrigin string, secureCookies bool, sessionTTL time.Duration) http.Handler {
	mux := http.NewServeMux()
	server := &Server{auth: authService, athlete: athleteService, onboarding: onboardingService, planning: planningService, secureCookies: secureCookies, sessionTTL: sessionTTL}
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
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("POST /v1/auth/register", server.register)
	mux.HandleFunc("POST /v1/auth/login", server.login)
	mux.HandleFunc("POST /v1/auth/logout", server.logout)
	mux.HandleFunc("GET /v1/me", server.me)
	mux.HandleFunc("GET /v1/profile", server.getProfile)
	mux.HandleFunc("PUT /v1/profile", server.putProfile)
	mux.HandleFunc("GET /v1/onboarding", server.getOnboarding)
	mux.HandleFunc("PUT /v1/onboarding/limitations", server.putLimitations)
	mux.HandleFunc("PUT /v1/onboarding/goals", server.putGoals)
	mux.HandleFunc("PUT /v1/onboarding/availability", server.putAvailability)
	mux.HandleFunc("POST /v1/plans/generate", server.generatePlan)
	mux.HandleFunc("GET /v1/plans/current", server.currentPlan)
	mux.HandleFunc("POST /v1/plans/{planID}/activate", server.activatePlan)
	mux.HandleFunc("POST /v1/workouts/{workoutID}/start", server.startWorkout)
	mux.HandleFunc("POST /v1/workouts/{workoutID}/complete", server.completeWorkout)
	mux.HandleFunc("POST /v1/workouts/{workoutID}/cancel", server.cancelWorkout)
	return cors(allowedOrigin, mux)
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
