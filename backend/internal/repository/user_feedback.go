package repository

import (
	"context"

	"github.com/Saulorangel87/App-de-treino/backend/internal/feedback"
)

func (s *Store) CreateUserFeedback(ctx context.Context, userID string, input feedback.CreateInput) (feedback.Entry, error) {
	var entry feedback.Entry
	err := s.pool.QueryRow(ctx, `
		INSERT INTO user_feedback (user_id, category, rating, message)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, user_id::text, category, rating, message, created_at`,
		userID, input.Category, input.Rating, input.Message,
	).Scan(&entry.ID, &entry.UserID, &entry.Category, &entry.Rating, &entry.Message, &entry.CreatedAt)
	return entry, err
}
