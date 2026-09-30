package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/activityimport"
	"github.com/Saulorangel87/App-de-treino/backend/internal/ai"
	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"github.com/Saulorangel87/App-de-treino/backend/internal/config"
	"github.com/Saulorangel87/App-de-treino/backend/internal/database"
	"github.com/Saulorangel87/App-de-treino/backend/internal/email"
	"github.com/Saulorangel87/App-de-treino/backend/internal/evolution"
	"github.com/Saulorangel87/App-de-treino/backend/internal/feedback"
	"github.com/Saulorangel87/App-de-treino/backend/internal/httpapi"
	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
	"github.com/Saulorangel87/App-de-treino/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// readiness makes /ready fail while migrations are missing.
type readiness struct {
	*pgxpool.Pool
	checker *database.SchemaChecker
}

func (r readiness) Check(ctx context.Context) error { return r.checker.Check(ctx) }

// purgeExpiredCredentials periodically removes expired sessions and e-mail
// tokens, which would otherwise accumulate forever.
func purgeExpiredCredentials(ctx context.Context, logger *slog.Logger, service *auth.Service) {
	run := func() {
		purgeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		sessions, tokens, err := service.PurgeExpired(purgeCtx)
		if err != nil {
			logger.Warn("expired credential purge failed", "error", err)
			return
		}
		if sessions+tokens > 0 {
			logger.Info("expired credentials purged", "sessions", sessions, "email_tokens", tokens)
		}
	}
	run()
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	db, err := database.Open(ctx, cfg.DatabaseURL, database.Options{MaxConns: cfg.DBMaxConns, MinConns: cfg.DBMinConns})
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// A healthy connection does not prove the schema matches this binary. In
	// production a mismatch is fatal, so a deploy that skipped migrations fails
	// loudly instead of serving 500s behind green health checks.
	schemaChecker := database.NewSchemaChecker(db)
	var ready httpapi.Pinger = db
	checkCtx, cancelCheck := context.WithTimeout(ctx, 10*time.Second)
	err = schemaChecker.Check(checkCtx)
	cancelCheck()
	if err != nil {
		if cfg.SecureCookies {
			logger.Error("database schema check failed", "error", err)
			os.Exit(1)
		}
		logger.Warn("database schema check failed; continuing because APP_ENV is not production", "error", err)
	}
	if cfg.SecureCookies {
		ready = readiness{Pool: db, checker: schemaChecker}
	}
	store := repository.New(db)
	authService := auth.NewService(store, cfg.SessionTTL)
	athleteService := athlete.NewService(store)
	onboardingService := athlete.NewOnboardingService(store)
	assessmentService := athlete.NewAssessmentService(store)
	recoveryService := athlete.NewRecoveryService(store)
	evolutionService := evolution.NewService(store)
	feedbackService := feedback.NewService(store)
	planningService := planning.NewService(store, planning.WithProtectionLevels(cfg.ProtectionLevelsEnabled))
	activityImportService := activityimport.NewService(store)
	var aiService *ai.Service
	if cfg.AIEnabled {
		var primary ai.Provider
		switch cfg.AIProvider {
		case "ollama":
			ollamaClient, err := ai.NewOllamaClient(cfg.AIBaseURL, cfg.AIModel, cfg.AITimeout, cfg.AIMaxTokens, cfg.AIMaxConcurrent)
			if err != nil {
				logger.Error("invalid Ollama configuration", "error", err)
				os.Exit(1)
			}
			primary = ollamaClient
		case "worker":
			if cfg.AIWorkerURL == "" || cfg.AIWorkerToken == "" {
				logger.Error("AI Worker provider requires AI_WORKER_URL and AI_WORKER_TOKEN")
				os.Exit(1)
			}
			workerClient, err := ai.NewWorkerClient(cfg.AIWorkerURL, cfg.AIWorkerToken, cfg.AITimeout, cfg.AIMaxConcurrent)
			if err != nil {
				logger.Error("invalid AI Worker configuration", "error", err)
				os.Exit(1)
			}
			primary = workerClient
		default:
			logger.Error("unsupported AI provider", "provider", cfg.AIProvider)
			os.Exit(1)
		}

		if cfg.AIProvider != "worker" && cfg.AIWorkerURL != "" {
			if cfg.AIWorkerToken == "" {
				logger.Warn("AI Worker fallback ignored because AI_WORKER_TOKEN is empty")
			} else {
				workerClient, err := ai.NewWorkerClient(cfg.AIWorkerURL, cfg.AIWorkerToken, cfg.AITimeout, cfg.AIMaxConcurrent)
				if err != nil {
					logger.Error("invalid AI Worker fallback configuration", "error", err)
					os.Exit(1)
				}
				primary = ai.NewFallbackProvider(primary, workerClient)
			}
		}
		aiService = ai.NewService(primary)
	} else {
		aiService = ai.NewService(nil)
	}
	var emailSender email.Sender = email.DevelopmentSender{}
	if cfg.ResendAPIKey != "" && cfg.EmailFrom != "" {
		emailSender = email.NewResendSender(cfg.ResendAPIKey, cfg.EmailFrom)
	} else if cfg.SecureCookies {
		emailSender = email.DisabledSender{}
	}

	// Explanations may wait on the AI provider, so the write deadline must
	// outlast its timeout.
	writeTimeout := max(30*time.Second, cfg.AITimeout+15*time.Second)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewRouter(ready, authService, athleteService, onboardingService, assessmentService, recoveryService, evolutionService, feedbackService, planningService, activityImportService, aiService, emailSender, cfg.AppBaseURL, cfg.AllowedOrigin, cfg.SecureCookies, cfg.Development, cfg.SessionTTL, cfg.EmailTokenTTL),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       60 * time.Second,
	}
	go purgeExpiredCredentials(ctx, logger, authService)
	go func() {
		logger.Info("api listening", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api stopped unexpectedly", "error", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
