package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"github.com/Saulorangel87/App-de-treino/backend/internal/email"
	"github.com/Saulorangel87/App-de-treino/backend/internal/legal"
)

type fakeLegalStore struct {
	accepted map[string]bool
	recorded []string
	err      error
}

func (f *fakeLegalStore) RecordLegalAcceptance(_ context.Context, userID, version string) error {
	if f.err != nil {
		return f.err
	}
	f.recorded = append(f.recorded, userID+"@"+version)
	if f.accepted == nil {
		f.accepted = map[string]bool{}
	}
	f.accepted[userID+"@"+version] = true
	return nil
}

func (f *fakeLegalStore) HasAcceptedLegal(_ context.Context, userID, version string) (bool, error) {
	return f.accepted[userID+"@"+version], nil
}

type registeringStore struct {
	httpTestAuthStore
	created int
}

func (s *registeringStore) CreateUser(_ context.Context, emailAddress, _, name string) (auth.User, error) {
	s.created++
	return auth.User{ID: "new-user", Email: emailAddress, DisplayName: name}, nil
}
func (s *registeringStore) CreateSession(context.Context, string, []byte, time.Time) error {
	return nil
}
func (s *registeringStore) CreateEmailToken(context.Context, string, string, []byte, time.Time) error {
	return nil
}

func registrationServer(legalStore *fakeLegalStore) (*Server, *registeringStore) {
	store := &registeringStore{}
	return &Server{
		auth: auth.NewService(store, time.Hour), sessionTTL: time.Hour, emailSender: email.DevelopmentSender{},
		appBaseURL: "http://localhost:3000", emailTokenTTL: time.Hour, development: true, legalAcceptances: legalStore,
	}, store
}

const registerBody = `{"email":"novo@example.com","password":"uma-senha-segura","display_name":"Novo"`

func TestRegisterRequiresAcceptingTheCurrentTerms(t *testing.T) {
	legalStore := &fakeLegalStore{}
	server, store := registrationServer(legalStore)

	for name, body := range map[string]string{
		"sem aceite":     registerBody + `}`,
		"aceite falso":   registerBody + `,"accept_terms":false,"terms_version":"` + legal.TermsVersion + `"}`,
		"versão ausente": registerBody + `,"accept_terms":true}`,
		"versão antiga":  registerBody + `,"accept_terms":true,"terms_version":"2020-01-01"}`,
	} {
		response := post(server.register, body, false)
		if response.Code != http.StatusBadRequest && response.Code != http.StatusConflict {
			t.Fatalf("%s: returned %d, want 400 or 409", name, response.Code)
		}
	}
	if store.created != 0 {
		t.Fatalf("%d account(s) were created without a valid acceptance", store.created)
	}
}

func TestRegisterRecordsTheAcceptedVersion(t *testing.T) {
	legalStore := &fakeLegalStore{}
	server, store := registrationServer(legalStore)

	response := post(server.register, registerBody+`,"accept_terms":true,"terms_version":"`+legal.TermsVersion+`"}`, false)
	if response.Code != http.StatusCreated {
		t.Fatalf("registration returned %d: %s", response.Code, response.Body)
	}
	if store.created != 1 || len(legalStore.recorded) != 1 || legalStore.recorded[0] != "new-user@"+legal.TermsVersion {
		t.Fatalf("created=%d recorded=%v, want one account with its acceptance", store.created, legalStore.recorded)
	}
}

func TestRegisterSurvivesAFailureToRecordTheAcceptance(t *testing.T) {
	server, _ := registrationServer(&fakeLegalStore{err: errors.New("db down")})
	response := post(server.register, registerBody+`,"accept_terms":true,"terms_version":"`+legal.TermsVersion+`"}`, false)
	if response.Code != http.StatusCreated {
		t.Fatalf("registration returned %d, want 201 (the app asks again on first access)", response.Code)
	}
}

func TestMeReportsWhetherTheCurrentTermsWereAccepted(t *testing.T) {
	store := httpTestAuthStore{user: auth.User{ID: "user-1"}}
	legalStore := &fakeLegalStore{}
	server := &Server{auth: auth.NewService(&store, time.Hour), legalAcceptances: legalStore}

	me := func() string {
		req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session-token"})
		response := httptest.NewRecorder()
		server.me(response, req)
		return response.Body.String()
	}
	if body := me(); !strings.Contains(body, `"accepted":false`) || !strings.Contains(body, `"terms_version":"`+legal.TermsVersion+`"`) {
		t.Fatalf("before accepting, /v1/me returned %s", body)
	}

	response := post(server.acceptTerms, `{"terms_version":"`+legal.TermsVersion+`"}`, true)
	if response.Code != http.StatusNoContent {
		t.Fatalf("accept returned %d, want 204", response.Code)
	}
	if body := me(); !strings.Contains(body, `"accepted":true`) {
		t.Fatalf("after accepting, /v1/me returned %s", body)
	}
}

func TestAcceptTermsRejectsAnOutdatedVersion(t *testing.T) {
	store := httpTestAuthStore{user: auth.User{ID: "user-1"}}
	legalStore := &fakeLegalStore{}
	server := &Server{auth: auth.NewService(&store, time.Hour), legalAcceptances: legalStore}

	if response := post(server.acceptTerms, `{"terms_version":"2020-01-01"}`, true); response.Code != http.StatusConflict {
		t.Fatalf("outdated version returned %d, want 409", response.Code)
	}
	if response := post(server.acceptTerms, `{"terms_version":"`+legal.TermsVersion+`"}`, false); response.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous acceptance returned %d, want 401", response.Code)
	}
	if len(legalStore.recorded) != 0 {
		t.Fatalf("recorded %v for invalid requests", legalStore.recorded)
	}
}
