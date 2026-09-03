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

	"github.com/Saulorangel87/App-de-treino/backend/internal/ai"
	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"github.com/Saulorangel87/App-de-treino/backend/internal/config"
	"github.com/Saulorangel87/App-de-treino/backend/internal/database"
	"github.com/Saulorangel87/App-de-treino/backend/internal/email"
	"github.com/Saulorangel87/App-de-treino/backend/internal/evolution"
	"github.com/Saulorangel87/App-de-treino/backend/internal/httpapi"
	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
	"github.com/Saulorangel87/App-de-treino/backend/internal/repository"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	store := repository.New(db)
	authService := auth.NewService(store, cfg.SessionTTL)
	athleteService := athlete.NewService(store)
	onboardingService := athlete.NewOnboardingService(store)
	assessmentService := athlete.NewAssessmentService(store)
	recoveryService := athlete.NewRecoveryService(store)
	evolutionService := evolution.NewService(store)
	planningService := planning.NewService(store)
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

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewRouter(db, authService, athleteService, onboardingService, assessmentService, recoveryService, evolutionService, planningService, aiService, emailSender, cfg.AppBaseURL, cfg.AllowedOrigin, cfg.SecureCookies, !cfg.SecureCookies, cfg.SessionTTL, cfg.EmailTokenTTL),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
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
