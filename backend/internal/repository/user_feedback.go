package repository

import (
	"context"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/feedback"
	"github.com/jackc/pgx/v5"
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

func (s *Store) PendingUserFeedback(ctx context.Context, limit int) ([]feedback.DigestEntry, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT uf.id::text, u.display_name, uf.category, uf.rating, uf.message, uf.created_at
		FROM user_feedback uf
		JOIN users u ON u.id = uf.user_id
		WHERE uf.digest_sent_at IS NULL
		ORDER BY uf.created_at ASC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	entries := make([]feedback.DigestEntry, 0)
	for rows.Next() {
		var entry feedback.DigestEntry
		if err := rows.Scan(&entry.ID, &entry.DisplayName, &entry.Category, &entry.Rating, &entry.Message, &entry.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *Store) MarkUserFeedbackDigested(ctx context.Context, ids []string, sentAt time.Time) error {
	return s.withTx(ctx, func(tx pgx.Tx) error {
		for _, id := range ids {
			if _, err := tx.Exec(ctx, `UPDATE user_feedback SET digest_sent_at = $1 WHERE id = $2 AND digest_sent_at IS NULL`, sentAt, id); err != nil {
				return err
			}
		}
		return nil
	})
}
