package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/config"
	"github.com/Saulorangel87/App-de-treino/backend/internal/database"
	"github.com/Saulorangel87/App-de-treino/backend/internal/email"
	"github.com/Saulorangel87/App-de-treino/backend/internal/feedback"
	"github.com/Saulorangel87/App-de-treino/backend/internal/repository"
)

const digestBatchSize = 50

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	if cfg.FeedbackDigestTo == "" {
		logger.Info("weekly feedback digest disabled: FEEDBACK_DIGEST_TO is empty")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	store := repository.New(db)
	entries, err := store.PendingUserFeedback(ctx, digestBatchSize)
	if err != nil {
		logger.Error("could not load pending feedback", "error", err)
		os.Exit(1)
	}
	if len(entries) == 0 {
		logger.Info("weekly feedback digest skipped: no new feedback")
		return
	}

	sender := email.NewResendSender(cfg.ResendAPIKey, cfg.EmailFrom)
	message := email.Message{
		To:      cfg.FeedbackDigestTo,
		Subject: feedback.DigestSubject(),
		HTML:    feedback.DigestHTML(entries),
		Text:    feedback.DigestText(entries),
	}
	if err := sender.Send(ctx, message); err != nil {
		logger.Error("feedback digest email failed", "error", err, "count", len(entries))
		os.Exit(1)
	}

	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.ID)
	}
	if err := store.MarkUserFeedbackDigested(ctx, ids, time.Now().UTC()); err != nil {
		logger.Error("feedback digest sent but could not mark entries", "error", err, "count", len(entries))
		os.Exit(1)
	}
	logger.Info("weekly feedback digest sent", "count", len(entries))
}
