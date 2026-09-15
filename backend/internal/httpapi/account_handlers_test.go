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

type accountDeletionStore struct {
	httpTestAuthStore
	deleted bool
}

func (s *accountDeletionStore) DeleteUser(_ context.Context, userID string) error {
	if s.user.ID != userID {
		return errHTTPTestUnused
	}
	s.deleted = true
	return nil
}

func TestDeleteAccountHandlerRequiresConfirmationAndPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("uma-senha-segura"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate password hash: %v", err)
	}
	store := &accountDeletionStore{httpTestAuthStore: httpTestAuthStore{user: auth.User{ID: "user-1", PasswordHash: string(hash)}}}
	server := &Server{auth: auth.NewService(store, time.Hour)}

	request := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodDelete, "/v1/auth/account", strings.NewReader(body))
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
		response := httptest.NewRecorder()
		server.deleteAccount(response, req)
		return response
	}

	if response := request(`{"password":"uma-senha-segura","confirmation":"APAGAR"}`); response.Code != http.StatusBadRequest {
		t.Fatalf("wrong confirmation returned %d, want 400", response.Code)
	}
	if store.deleted {
		t.Fatal("account was deleted with a wrong confirmation")
	}

	if response := request(`{"password":"senha-incorreta","confirmation":"ENCERRAR CONTA"}`); response.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password returned %d, want 401", response.Code)
	}
	if store.deleted {
		t.Fatal("account was deleted with a wrong password")
	}

	response := request(`{"password":"uma-senha-segura","confirmation":"ENCERRAR CONTA"}`)
	if response.Code != http.StatusNoContent || !store.deleted {
		t.Fatalf("valid deletion returned %d with deleted=%t, want 204 and true", response.Code, store.deleted)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != sessionCookieName || cookies[0].MaxAge != -1 {
		t.Fatalf("expected expired session cookie, got %#v", cookies)
	}
}
