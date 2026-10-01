package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
)

type fakeExporter struct {
	userID   string
	sections map[string]json.RawMessage
	err      error
}

func (f *fakeExporter) ExportAccountData(_ context.Context, userID string) (map[string]json.RawMessage, error) {
	f.userID = userID
	return f.sections, f.err
}

func exportRequest(server *Server) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/account/export", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	response := httptest.NewRecorder()
	server.exportAccountData(response, req)
	return response
}

func TestExportAccountDataReturnsOnlyTheSignedInUsersSections(t *testing.T) {
	store := httpTestAuthStore{user: auth.User{ID: "user-1"}}
	exporter := &fakeExporter{sections: map[string]json.RawMessage{
		"account": json.RawMessage(`{"email":"a@example.invalid"}`),
		"goals":   json.RawMessage(`[]`),
	}}
	server := &Server{auth: auth.NewService(&store, time.Hour), exporter: exporter}

	response := exportRequest(server)
	if response.Code != http.StatusOK {
		t.Fatalf("export returned %d, want 200: %s", response.Code, response.Body)
	}
	if exporter.userID != "user-1" {
		t.Fatalf("exported data for %q, want the signed-in user", exporter.userID)
	}
	if disposition := response.Header().Get("Content-Disposition"); !strings.HasPrefix(disposition, `attachment; filename="cadencia-dados-`) {
		t.Fatalf("unexpected content disposition %q", disposition)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("the export must not be cached")
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(response.Body.Bytes(), &document); err != nil {
		t.Fatalf("export is not valid JSON: %v", err)
	}
	for _, key := range []string{"format_version", "exported_at", "notice", "account", "goals"} {
		if _, ok := document[key]; !ok {
			t.Fatalf("export is missing %q", key)
		}
	}
	for _, forbidden := range []string{"password", "token_hash"} {
		if strings.Contains(response.Body.String(), forbidden) {
			t.Fatalf("export must not mention %q", forbidden)
		}
	}
}

func TestExportAccountDataFailsWithoutLeakingDetails(t *testing.T) {
	store := httpTestAuthStore{user: auth.User{ID: "user-1"}}
	server := &Server{auth: auth.NewService(&store, time.Hour), exporter: &fakeExporter{err: errors.New("pq: relation secret_table does not exist")}}

	response := exportRequest(server)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("failed export returned %d, want 500", response.Code)
	}
	if strings.Contains(response.Body.String(), "secret_table") {
		t.Fatal("the internal error leaked to the client")
	}
}

func TestExportAccountDataRequiresSession(t *testing.T) {
	server := &Server{auth: auth.NewService(&httpTestAuthStore{}, time.Hour), exporter: &fakeExporter{}}
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/account/export", nil)
	response := httptest.NewRecorder()
	server.exportAccountData(response, req)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous export returned %d, want 401", response.Code)
	}
}
