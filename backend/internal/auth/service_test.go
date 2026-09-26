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
	deleted     bool

	otherSessionsRevoked bool
	keptSession          []byte
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
func (s *memoryStore) DeleteUser(_ context.Context, userID string) error {
	if s.user.ID != userID {
		return errors.New("not found")
	}
	s.deleted = true
	s.user = User{}
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

func (s *memoryStore) ChangePassword(_ context.Context, userID, passwordHash string, keepSessionHash []byte) error {
	if s.user.ID != userID {
		return errors.New("not found")
	}
	s.user.PasswordHash = passwordHash
	s.otherSessionsRevoked = true
	s.keptSession = append([]byte(nil), keepSessionHash...)
	return nil
}
func (s *memoryStore) DeleteOtherSessions(_ context.Context, userID string, keepSessionHash []byte) (int64, error) {
	if s.user.ID != userID {
		return 0, errors.New("not found")
	}
	s.otherSessionsRevoked = true
	s.keptSession = append([]byte(nil), keepSessionHash...)
	return 2, nil
}
func (s *memoryStore) PurgeExpired(context.Context) (int64, int64, error) { return 1, 2, nil }

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

func TestDeleteAccountRequiresCurrentPassword(t *testing.T) {
	store := &memoryStore{}
	service := NewService(store, time.Hour)
	user, _, err := service.Register(context.Background(), "atleta@example.com", "uma-senha-segura", "Atleta")
	if err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}
	if err := service.DeleteAccount(context.Background(), user, "senha-incorreta"); err != ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	if store.deleted {
		t.Fatal("account was deleted with an invalid password")
	}
	if err := service.DeleteAccount(context.Background(), user, "uma-senha-segura"); err != nil {
		t.Fatalf("unexpected account deletion error: %v", err)
	}
	if !store.deleted {
		t.Fatal("account was not deleted after password confirmation")
	}
}

func TestLoginUnknownEmailSpendsBcryptCost(t *testing.T) {
	service := NewService(&memoryStore{}, time.Hour)
	// Warm up so the first bcrypt call does not skew the comparison.
	_, _, _ = service.Login(context.Background(), "ninguem@example.com", "qualquer-senha")
	started := time.Now()
	if _, _, err := service.Login(context.Background(), "ninguem@example.com", "qualquer-senha"); err != ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	// bcrypt at the default cost takes tens of milliseconds; returning without
	// hashing would take microseconds.
	if elapsed := time.Since(started); elapsed < 5*time.Millisecond {
		t.Fatalf("unknown e-mail login returned in %s, expected bcrypt-scale work", elapsed)
	}
}

func TestChangePasswordVerifiesCurrentPasswordAndKeepsCurrentSession(t *testing.T) {
	store := &memoryStore{}
	service := NewService(store, time.Hour)
	user, token, err := service.Register(context.Background(), "atleta@example.com", "uma-senha-segura", "Atleta")
	if err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}
	if err := service.ChangePassword(context.Background(), user, token, "senha-incorreta", "nova-senha-segura"); err != ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	if err := service.ChangePassword(context.Background(), user, token, "uma-senha-segura", "curta"); err != ErrInvalidInput {
		t.Fatalf("expected invalid input for a short password, got %v", err)
	}
	if err := service.ChangePassword(context.Background(), user, token, "uma-senha-segura", "uma-senha-segura"); err != ErrInvalidInput {
		t.Fatalf("expected invalid input when the password does not change, got %v", err)
	}
	if store.otherSessionsRevoked {
		t.Fatal("sessions were revoked by a rejected change")
	}
	if err := service.ChangePassword(context.Background(), user, token, "uma-senha-segura", "nova-senha-segura"); err != nil {
		t.Fatalf("unexpected change error: %v", err)
	}
	if !store.otherSessionsRevoked || !bytes.Equal(store.keptSession, hashToken(token)) {
		t.Fatal("expected other sessions revoked while keeping the current one")
	}
	if _, _, err := service.Login(context.Background(), user.Email, "nova-senha-segura"); err != nil {
		t.Fatalf("expected login with the new password: %v", err)
	}
}

func TestLogoutOthersRequiresSession(t *testing.T) {
	store := &memoryStore{}
	service := NewService(store, time.Hour)
	user, token, _ := service.Register(context.Background(), "atleta@example.com", "uma-senha-segura", "Atleta")
	if _, err := service.LogoutOthers(context.Background(), user, ""); err != ErrUnauthorized {
		t.Fatalf("expected unauthorized without a session token, got %v", err)
	}
	revoked, err := service.LogoutOthers(context.Background(), user, token)
	if err != nil || revoked != 2 {
		t.Fatalf("expected 2 revoked sessions, got %d / %v", revoked, err)
	}
}
