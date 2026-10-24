package httpapi

import (
	"context"
	"fmt"
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
	server := &Server{auth: auth.NewService(store, time.Hour), sessionTTL: time.Hour, loginAttempts: newMemoryLoginAttempts()}
	server.now = func() time.Time { return testNow }
	server.wait = func(context.Context, time.Duration) {}
	return server, store
}

var testNow = time.Date(2026, time.October, 23, 21, 0, 0, 0, time.UTC)

func post(handler http.HandlerFunc, body string, withSession bool) *httptest.ResponseRecorder {
	return postFrom(handler, body, withSession, "192.0.2.10:1234")
}

func postFrom(handler http.HandlerFunc, body string, withSession bool, remoteAddr string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.RemoteAddr = remoteAddr
	if withSession {
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session-token"})
	}
	response := httptest.NewRecorder()
	handler(response, req)
	return response
}

const (
	wrongLogin = `{"email":"atleta@example.com","password":"senha-errada-123"}`
	rightLogin = `{"email":"Atleta@Example.com","password":"uma-senha-segura"}`
)

func TestLoginWarnsOnLastAttemptAndBlocksTheIPForTenMinutes(t *testing.T) {
	server, _ := newSessionServer(t)
	for attempt := 1; attempt <= 3; attempt++ {
		response := post(server.login, wrongLogin, false)
		if response.Code != http.StatusUnauthorized || strings.Contains(response.Body.String(), "resta 1") {
			t.Fatalf("attempt %d: got %d %s, want a plain 401", attempt, response.Code, response.Body.String())
		}
	}
	warning := post(server.login, wrongLogin, false)
	if warning.Code != http.StatusUnauthorized || !strings.Contains(warning.Body.String(), "resta 1 tentativa") {
		t.Fatalf("4th attempt: got %d %s, want the last-attempt warning", warning.Code, warning.Body.String())
	}
	blocked := post(server.login, wrongLogin, false)
	if blocked.Code != http.StatusTooManyRequests || blocked.Header().Get("Retry-After") != "600" || !strings.Contains(blocked.Body.String(), "login_blocked") {
		t.Fatalf("5th attempt: got %d Retry-After %q %s, want a 10-minute block", blocked.Code, blocked.Header().Get("Retry-After"), blocked.Body.String())
	}
	// Even the right password is refused from that IP while blocked...
	if code := post(server.login, rightLogin, false).Code; code != http.StatusTooManyRequests {
		t.Fatalf("right password while blocked returned %d, want 429", code)
	}
	// ...but another IP signs in normally: the account itself is never locked.
	if code := postFrom(server.login, rightLogin, false, "198.51.100.7:4321").Code; code != http.StatusOK {
		t.Fatalf("other IP returned %d, want 200", code)
	}
	// After 10 minutes the IP gets a fresh count of 5.
	server.now = func() time.Time { return testNow.Add(loginBlockDuration) }
	if code := post(server.login, rightLogin, false).Code; code != http.StatusOK {
		t.Fatalf("after the block returned %d, want 200", code)
	}
	for attempt := 1; attempt <= 3; attempt++ {
		if response := post(server.login, wrongLogin, false); strings.Contains(response.Body.String(), "resta 1") {
			t.Fatalf("attempt %d after the block already warned", attempt)
		}
	}
}

func TestLoginErrorsFarApartDoNotAddUp(t *testing.T) {
	server, _ := newSessionServer(t)
	for attempt := 0; attempt < 8; attempt++ {
		at := testNow.Add(time.Duration(attempt) * (loginFailureWindow + time.Second))
		server.now = func() time.Time { return at }
		if code := post(server.login, wrongLogin, false).Code; code != http.StatusUnauthorized {
			t.Fatalf("spaced attempt %d returned %d, want 401", attempt+1, code)
		}
	}
}

func TestSuccessfulLoginDoesNotResetTheIPCount(t *testing.T) {
	server, _ := newSessionServer(t)
	for attempt := 0; attempt < 3; attempt++ {
		post(server.login, wrongLogin, false)
	}
	if code := post(server.login, rightLogin, false).Code; code != http.StatusOK {
		t.Fatalf("valid login returned %d, want 200", code)
	}
	if response := post(server.login, wrongLogin, false); !strings.Contains(response.Body.String(), "resta 1 tentativa") {
		t.Fatalf("a correct login must not give the IP new attempts, got %s", response.Body.String())
	}
}

func TestAccountUnderAttackFromManyIPsGetsSlowerButNotLocked(t *testing.T) {
	server, _ := newSessionServer(t)
	var delays []time.Duration
	server.wait = func(_ context.Context, delay time.Duration) { delays = append(delays, delay) }
	for attempt := 0; attempt < 12; attempt++ {
		postFrom(server.login, wrongLogin, false, fmt.Sprintf("203.0.113.%d:1000", attempt+1))
	}
	if len(delays) == 0 || delays[len(delays)-1] != 3*time.Second {
		t.Fatalf("delays = %v, want growing waits ending at 3s", delays)
	}
	if code := postFrom(server.login, rightLogin, false, "198.51.100.7:4321").Code; code != http.StatusOK {
		t.Fatalf("owner returned %d, want 200 (the account must not be locked)", code)
	}
	delays = nil
	postFrom(server.login, wrongLogin, false, "203.0.113.200:1000")
	if len(delays) != 0 {
		t.Fatalf("the owner's login should clear the slowdown, got %v", delays)
	}
}

func TestLoginKeysDoNotStoreAddressesInClearText(t *testing.T) {
	key := loginKey("ip", "198.51.100.7")
	if strings.Contains(key, "198.51") || len(key) != len("ip:")+64 {
		t.Fatalf("key %q must be a SHA-256 digest", key)
	}
	if loginKey("account", " Atleta@Example.com ") != loginKey("account", "atleta@example.com") {
		t.Fatal("account keys must ignore case and surrounding spaces")
	}
}

func TestMemoryLoginAttemptsDropsTheOldestKeyWhenFull(t *testing.T) {
	store := &memoryLoginAttempts{entries: map[string]auth.LoginFailures{}, maxKeys: 2}
	ctx := context.Background()
	for index, key := range []string{"old", "mid", "new"} {
		at := testNow.Add(time.Duration(index) * time.Minute)
		if _, err := store.UpdateLoginFailures(ctx, key, func(state auth.LoginFailures) auth.LoginFailures {
			return countFailure(state, at, 0, time.Hour, 0)
		}); err != nil {
			t.Fatal(err)
		}
	}
	if _, ok := store.entries["old"]; ok || len(store.entries) != 2 {
		t.Fatalf("entries = %v, want the oldest dropped", store.entries)
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
