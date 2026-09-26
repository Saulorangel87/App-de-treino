package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

type sessionStore struct {
	httpTestAuthStore
	changedTo    string
	keptSession  []byte
	otherRevoked bool
}

func (s *sessionStore) UserByEmail(_ context.Context, email string) (auth.User, error) {
	if email != s.user.Email {
		return auth.User{}, errHTTPTestUnused
	}
	return s.user, nil
}
func (s *sessionStore) CreateSession(context.Context, string, []byte, time.Time) error { return nil }
func (s *sessionStore) ChangePassword(_ context.Context, _ string, hash string, keep []byte) error {
	s.changedTo, s.keptSession = hash, keep
	return nil
}
func (s *sessionStore) DeleteOtherSessions(_ context.Context, _ string, keep []byte) (int64, error) {
	s.keptSession, s.otherRevoked = keep, true
	return 3, nil
}

func newSessionServer(t *testing.T) (*Server, *sessionStore) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("uma-senha-segura"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	store := &sessionStore{httpTestAuthStore: httpTestAuthStore{user: auth.User{ID: "user-1", Email: "atleta@example.com", PasswordHash: string(hash)}}}
	return &Server{auth: auth.NewService(store, time.Hour), sessionTTL: time.Hour, loginFailures: newRequestRateLimiter()}, store
}

func post(handler http.HandlerFunc, body string, withSession bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	if withSession {
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session-token"})
	}
	response := httptest.NewRecorder()
	handler(response, req)
	return response
}

func TestLoginBlocksAccountAfterRepeatedFailuresAndResetsOnSuccess(t *testing.T) {
	server, _ := newSessionServer(t)
	wrong := `{"email":"atleta@example.com","password":"senha-errada-123"}`
	for i := 0; i < maxLoginFailuresPerAccount; i++ {
		if code := post(server.login, wrong, false).Code; code != http.StatusUnauthorized {
			t.Fatalf("attempt %d returned %d, want 401", i+1, code)
		}
	}
	blocked := post(server.login, `{"email":"Atleta@Example.com","password":"uma-senha-segura"}`, false)
	if blocked.Code != http.StatusTooManyRequests || blocked.Header().Get("Retry-After") == "" {
		t.Fatalf("expected 429 with Retry-After after repeated failures, got %d", blocked.Code)
	}

	server, _ = newSessionServer(t)
	for i := 0; i < maxLoginFailuresPerAccount-1; i++ {
		post(server.login, wrong, false)
	}
	if code := post(server.login, `{"email":"atleta@example.com","password":"uma-senha-segura"}`, false).Code; code != http.StatusOK {
		t.Fatalf("valid login returned %d, want 200", code)
	}
	if blocked, _ := server.loginFailures.exceeded("login-account:atleta@example.com", 1); blocked {
		t.Fatal("a successful login should clear the failure counter")
	}
}

func TestChangePasswordHandler(t *testing.T) {
	server, store := newSessionServer(t)
	if code := post(server.changePassword, `{"current_password":"uma-senha-segura","new_password":"nova-senha-segura"}`, false).Code; code != http.StatusUnauthorized {
		t.Fatalf("without a session got %d, want 401", code)
	}
	if code := post(server.changePassword, `{"current_password":"errada-errada","new_password":"nova-senha-segura"}`, true).Code; code != http.StatusUnauthorized {
		t.Fatalf("wrong current password got %d, want 401", code)
	}
	if code := post(server.changePassword, `{"current_password":"uma-senha-segura","new_password":"curta"}`, true).Code; code != http.StatusBadRequest {
		t.Fatalf("short new password got %d, want 400", code)
	}
	if store.changedTo != "" {
		t.Fatal("password changed despite rejected requests")
	}
	if code := post(server.changePassword, `{"current_password":"uma-senha-segura","new_password":"nova-senha-segura"}`, true).Code; code != http.StatusOK {
		t.Fatalf("valid change got %d, want 200", code)
	}
	if store.changedTo == "" || len(store.keptSession) == 0 {
		t.Fatal("expected the new hash stored while keeping the current session")
	}
}

func TestLogoutOthersHandler(t *testing.T) {
	server, store := newSessionServer(t)
	if code := post(server.logoutOthers, ``, false).Code; code != http.StatusUnauthorized {
		t.Fatalf("without a session got %d, want 401", code)
	}
	response := post(server.logoutOthers, ``, true)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"revoked_sessions":3`) || !store.otherRevoked {
		t.Fatalf("unexpected response %d %s", response.Code, response.Body.String())
	}
}
