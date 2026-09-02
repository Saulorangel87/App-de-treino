package auth

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

type memoryStore struct {
	user        User
	tokenHash   []byte
	expiresAt   time.Time
	emailTokens map[string][]byte
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
func (s *memoryStore) CreateEmailToken(_ context.Context, _ string, purpose string, tokenHash []byte, _ time.Time) error {
	if s.emailTokens == nil {
		s.emailTokens = map[string][]byte{}
	}
	s.emailTokens[purpose] = append([]byte(nil), tokenHash...)
	return nil
}
func (s *memoryStore) VerifyEmailToken(_ context.Context, tokenHash []byte) (User, error) {
	if !bytes.Equal(s.emailTokens["verify_email"], tokenHash) {
		return User{}, errors.New("not found")
	}
	s.emailTokens["verify_email"] = nil
	s.user.EmailVerified = true
	return s.user, nil
}
func (s *memoryStore) ResetPasswordWithToken(_ context.Context, tokenHash []byte, passwordHash string) error {
	if !bytes.Equal(s.emailTokens["reset_password"], tokenHash) {
		return errors.New("not found")
	}
	s.emailTokens["reset_password"] = nil
	s.user.PasswordHash = passwordHash
	s.tokenHash = nil
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

func TestVerifyEmailAndResetPasswordUseSingleUseTokens(t *testing.T) {
	store := &memoryStore{}
	service := NewService(store, time.Hour)
	user, _, err := service.Register(context.Background(), "atleta@example.com", "uma-senha-segura", "Atleta")
	if err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}
	verification, err := service.CreateEmailVerificationToken(context.Background(), user.ID, time.Hour)
	if err != nil {
		t.Fatalf("unexpected token error: %v", err)
	}
	verified, err := service.VerifyEmail(context.Background(), verification)
	if err != nil || !verified.EmailVerified {
		t.Fatalf("expected verified user, got %#v / %v", verified, err)
	}
	if _, err := service.VerifyEmail(context.Background(), verification); err != ErrInvalidEmailToken {
		t.Fatalf("expected consumed token to fail, got %v", err)
	}
	_, reset, err := service.CreatePasswordResetToken(context.Background(), user.Email, time.Hour)
	if err != nil || reset == "" {
		t.Fatalf("expected reset token, got %q / %v", reset, err)
	}
	if err := service.ResetPassword(context.Background(), reset, "nova-senha-segura"); err != nil {
		t.Fatalf("unexpected reset error: %v", err)
	}
	if _, _, err := service.Login(context.Background(), user.Email, "nova-senha-segura"); err != nil {
		t.Fatalf("expected login with new password: %v", err)
	}
}
