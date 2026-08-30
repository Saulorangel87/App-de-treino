package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already registered")
	ErrUnauthorized       = errors.New("unauthorized")
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Store interface {
	CreateUser(context.Context, string, string, string) (User, error)
	UserByEmail(context.Context, string) (User, error)
	CreateSession(context.Context, string, []byte, time.Time) error
	UserBySessionHash(context.Context, []byte) (User, error)
	DeleteSession(context.Context, []byte) error
}

type Service struct {
	store Store
	ttl   time.Duration
	now   func() time.Time
}

func NewService(store Store, ttl time.Duration) *Service {
	return &Service{store: store, ttl: ttl, now: time.Now}
}

func (s *Service) Register(ctx context.Context, email, password, displayName string) (User, string, error) {
	email, displayName, err := validateRegistration(email, password, displayName)
	if err != nil {
		return User{}, "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, "", err
	}
	user, err := s.store.CreateUser(ctx, email, string(hash), displayName)
	if err != nil {
		return User{}, "", err
	}
	token, err := s.newSession(ctx, user.ID)
	return user, token, err
}

func (s *Service) Login(ctx context.Context, email, password string) (User, string, error) {
	user, err := s.store.UserByEmail(ctx, normalizeEmail(email))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return User{}, "", ErrInvalidCredentials
	}
	token, err := s.newSession(ctx, user.ID)
	return user, token, err
}

func (s *Service) Authenticate(ctx context.Context, token string) (User, error) {
	if token == "" {
		return User{}, ErrUnauthorized
	}
	user, err := s.store.UserBySessionHash(ctx, hashToken(token))
	if err != nil {
		return User{}, ErrUnauthorized
	}
	return user, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.store.DeleteSession(ctx, hashToken(token))
}

func (s *Service) newSession(ctx context.Context, userID string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	if err := s.store.CreateSession(ctx, userID, hashToken(token), s.now().Add(s.ttl)); err != nil {
		return "", err
	}
	return token, nil
}

func hashToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}

func validateRegistration(email, password, displayName string) (string, string, error) {
	email = normalizeEmail(email)
	displayName = strings.TrimSpace(displayName)
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > 254 || len(displayName) < 2 || len(displayName) > 100 || len(password) < 10 || len(password) > 72 {
		return "", "", ErrInvalidInput
	}
	return email, displayName, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
