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

	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"github.com/Saulorangel87/App-de-treino/backend/internal/config"
	"github.com/Saulorangel87/App-de-treino/backend/internal/database"
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
	planningService := planning.NewService(store)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewRouter(db, authService, athleteService, onboardingService, planningService, cfg.AllowedOrigin, cfg.SecureCookies, cfg.SessionTTL),
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
