package repository

import (
	"context"
	"sync"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/feedback"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const feedbackDigestLockKey int64 = 6_943_150_214

type feedbackQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

// FeedbackDigestLock keeps the pending read and sent marker in one database
// transaction, preventing the timer and a manual run from processing the same
// batch concurrently.
type FeedbackDigestLock struct {
	conn   *pgxpool.Conn
	tx     pgx.Tx
	mu     sync.Mutex
	closed bool
}

func (s *Store) AcquireFeedbackDigestLock(ctx context.Context) (*FeedbackDigestLock, bool, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, false, err
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		conn.Release()
		return nil, false, err
	}
	var acquired bool
	if err := tx.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock($1)`, feedbackDigestLockKey).Scan(&acquired); err != nil {
		_ = tx.Rollback(ctx)
		conn.Release()
		return nil, false, err
	}
	if !acquired {
		_ = tx.Rollback(ctx)
		conn.Release()
		return nil, false, nil
	}
	return &FeedbackDigestLock{conn: conn, tx: tx}, true, nil
}

func (l *FeedbackDigestLock) PendingUserFeedback(ctx context.Context, limit int) ([]feedback.DigestEntry, error) {
	return pendingUserFeedback(ctx, l.tx, limit)
}

func (l *FeedbackDigestLock) MarkUserFeedbackDigested(ctx context.Context, ids []string, sentAt time.Time) error {
	return markUserFeedbackDigested(ctx, l.tx, ids, sentAt)
}

func (l *FeedbackDigestLock) Commit(ctx context.Context) error {
	return l.finish(ctx, true)
}

func (l *FeedbackDigestLock) Rollback(ctx context.Context) error {
	return l.finish(ctx, false)
}

func (l *FeedbackDigestLock) finish(ctx context.Context, commit bool) error {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return nil
	}
	l.closed = true
	l.mu.Unlock()

	var err error
	if commit {
		err = l.tx.Commit(ctx)
	} else {
		err = l.tx.Rollback(ctx)
	}
	l.conn.Release()
	return err
}

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
	return pendingUserFeedback(ctx, s.pool, limit)
}

func pendingUserFeedback(ctx context.Context, queryer feedbackQueryer, limit int) ([]feedback.DigestEntry, error) {
	rows, err := queryer.Query(ctx, `
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
		return markUserFeedbackDigested(ctx, tx, ids, sentAt)
	})
}

func markUserFeedbackDigested(ctx context.Context, tx pgx.Tx, ids []string, sentAt time.Time) error {
	for _, id := range ids {
		if _, err := tx.Exec(ctx, `UPDATE user_feedback SET digest_sent_at = $1 WHERE id = $2 AND digest_sent_at IS NULL`, sentAt, id); err != nil {
			return err
		}
	}
	return nil
}
