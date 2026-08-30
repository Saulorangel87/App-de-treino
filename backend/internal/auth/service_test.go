package auth

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

type memoryStore struct {
	user      User
	tokenHash []byte
	expiresAt time.Time
}

func (s *memoryStore) CreateUser(_ context.Context, email, passwordHash, displayName string) (User, error) {
	s.user = User{ID: "user-1", Email: email, PasswordHash: passwordHash, DisplayName: displayName, CreatedAt: time.Now()}
	return s.user, nil
}
func (s *memoryStore) UserByEmail(_ context.Context, email string) (User, error) {
	if s.user.Email != email {
		return User{}, errors.New("not found")
	}
	return s.user, nil
}
func (s *memoryStore) CreateSession(_ context.Context, _ string, tokenHash []byte, expiresAt time.Time) error {
	s.tokenHash = append([]byte(nil), tokenHash...)
	s.expiresAt = expiresAt
	return nil
}
func (s *memoryStore) UserBySessionHash(_ context.Context, tokenHash []byte) (User, error) {
	if !bytes.Equal(s.tokenHash, tokenHash) {
		return User{}, errors.New("not found")
	}
	return s.user, nil
}
func (s *memoryStore) DeleteSession(_ context.Context, tokenHash []byte) error {
	if bytes.Equal(s.tokenHash, tokenHash) {
		s.tokenHash = nil
	}
	return nil
}

func TestRegisterCreatesHashedPasswordAndSession(t *testing.T) {
	store := &memoryStore{}
	service := NewService(store, 7*24*time.Hour)
	user, token, err := service.Register(context.Background(), " ATLETA@EXAMPLE.COM ", "uma-senha-segura", "Atleta Teste")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "atleta@example.com" || token == "" || store.user.PasswordHash == "uma-senha-segura" {
		t.Fatal("registration did not normalize email, hash password and create session")
	}
	if _, err := service.Authenticate(context.Background(), token); err != nil {
		t.Fatalf("expected session to authenticate: %v", err)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	store := &memoryStore{}
	service := NewService(store, time.Hour)
	_, _, _ = service.Register(context.Background(), "atleta@example.com", "uma-senha-segura", "Atleta")
	if _, _, err := service.Login(context.Background(), "atleta@example.com", "senha-errada"); err != ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}
