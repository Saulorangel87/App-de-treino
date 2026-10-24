package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
)

// staleLoginFailures is how long a key without new errors stays stored.
const staleLoginFailures = 24 * time.Hour

func (s *Store) LoginFailures(ctx context.Context, key string) (auth.LoginFailures, error) {
	var state auth.LoginFailures
	var blockedUntil *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT failures, last_failure_at, blocked_until FROM login_failures WHERE key = $1`, key).
		Scan(&state.Failures, &state.LastFailureAt, &blockedUntil)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.LoginFailures{}, nil
	}
	if blockedUntil != nil {
		state.BlockedUntil = *blockedUntil
	}
	return state, err
}

// UpdateLoginFailures applies update to the stored state under a row lock, so
// two wrong passwords sent at the same time are both counted. It also removes
// keys that have had no errors for a day.
func (s *Store) UpdateLoginFailures(ctx context.Context, key string, update func(auth.LoginFailures) auth.LoginFailures) (auth.LoginFailures, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return auth.LoginFailures{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		DELETE FROM login_failures
		WHERE last_failure_at < now() - make_interval(secs => $1)
		  AND (blocked_until IS NULL OR blocked_until < now())`, staleLoginFailures.Seconds()); err != nil {
		return auth.LoginFailures{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO login_failures (key) VALUES ($1) ON CONFLICT (key) DO NOTHING`, key); err != nil {
		return auth.LoginFailures{}, err
	}
	var state auth.LoginFailures
	var blockedUntil *time.Time
	if err := tx.QueryRow(ctx, `
		SELECT failures, last_failure_at, blocked_until FROM login_failures WHERE key = $1 FOR UPDATE`, key).
		Scan(&state.Failures, &state.LastFailureAt, &blockedUntil); err != nil {
		return auth.LoginFailures{}, err
	}
	if blockedUntil != nil {
		state.BlockedUntil = *blockedUntil
	}
	state = update(state)
	var stored *time.Time
	if !state.BlockedUntil.IsZero() {
		stored = &state.BlockedUntil
	}
	if _, err := tx.Exec(ctx, `
		UPDATE login_failures SET failures = $2, last_failure_at = $3, blocked_until = $4 WHERE key = $1`,
		key, state.Failures, state.LastFailureAt, stored); err != nil {
		return auth.LoginFailures{}, err
	}
	return state, tx.Commit(ctx)
}

func (s *Store) ClearLoginFailures(ctx context.Context, key string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM login_failures WHERE key = $1`, key)
	return err
}
