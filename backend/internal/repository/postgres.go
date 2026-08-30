package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) CreateUser(ctx context.Context, email, passwordHash, displayName string) (auth.User, error) {
	var user auth.User
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name)
		VALUES ($1, $2, $3)
		RETURNING id::text, email, password_hash, display_name, created_at`, email, passwordHash, displayName,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return auth.User{}, auth.ErrEmailExists
	}
	return user, err
}

func (s *Store) UserByEmail(ctx context.Context, email string) (auth.User, error) {
	var user auth.User
	err := s.pool.QueryRow(ctx, `SELECT id::text, email, password_hash, display_name, created_at FROM users WHERE email = $1`, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.CreatedAt)
	return user, err
}

func (s *Store) CreateSession(ctx context.Context, userID string, tokenHash []byte, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO auth_sessions (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`, userID, tokenHash, expiresAt)
	return err
}

func (s *Store) UserBySessionHash(ctx context.Context, tokenHash []byte) (auth.User, error) {
	var user auth.User
	err := s.pool.QueryRow(ctx, `
		UPDATE auth_sessions SET last_used_at = now()
		WHERE token_hash = $1 AND expires_at > now()
		RETURNING (SELECT id::text FROM users WHERE id = auth_sessions.user_id),
		          (SELECT email FROM users WHERE id = auth_sessions.user_id),
		          (SELECT password_hash FROM users WHERE id = auth_sessions.user_id),
		          (SELECT display_name FROM users WHERE id = auth_sessions.user_id),
		          (SELECT created_at FROM users WHERE id = auth_sessions.user_id)`, tokenHash,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.CreatedAt)
	return user, err
}

func (s *Store) DeleteSession(ctx context.Context, tokenHash []byte) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM auth_sessions WHERE token_hash = $1`, tokenHash)
	return err
}

func (s *Store) UpsertProfile(ctx context.Context, profile athlete.Profile) (athlete.Profile, error) {
	var saved athlete.Profile
	err := s.pool.QueryRow(ctx, `
		INSERT INTO athlete_profiles (user_id, birth_date, sex, height_cm, weight_kg, sport, experience_level, activity_level)
		VALUES ($1, $2::date, $3, $4, $5, 'cycling', $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			birth_date = EXCLUDED.birth_date, sex = EXCLUDED.sex, height_cm = EXCLUDED.height_cm,
			weight_kg = EXCLUDED.weight_kg, experience_level = EXCLUDED.experience_level,
			activity_level = EXCLUDED.activity_level, updated_at = now()
		RETURNING user_id::text, birth_date::text, sex, height_cm::double precision,
			weight_kg::double precision, sport, experience_level, activity_level`,
		profile.UserID, profile.BirthDate, profile.Sex, profile.HeightCM, profile.WeightKG, profile.ExperienceLevel, profile.ActivityLevel,
	).Scan(&saved.UserID, &saved.BirthDate, &saved.Sex, &saved.HeightCM, &saved.WeightKG, &saved.Sport, &saved.ExperienceLevel, &saved.ActivityLevel)
	return saved, err
}

func (s *Store) ProfileByUserID(ctx context.Context, userID string) (athlete.Profile, error) {
	var profile athlete.Profile
	err := s.pool.QueryRow(ctx, `
		SELECT user_id::text, birth_date::text, sex, height_cm::double precision,
			weight_kg::double precision, sport, experience_level, activity_level
		FROM athlete_profiles WHERE user_id = $1`, userID,
	).Scan(&profile.UserID, &profile.BirthDate, &profile.Sex, &profile.HeightCM, &profile.WeightKG, &profile.Sport, &profile.ExperienceLevel, &profile.ActivityLevel)
	if errors.Is(err, pgx.ErrNoRows) {
		return athlete.Profile{}, athlete.ErrProfileMissing
	}
	return profile, err
}
