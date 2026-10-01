package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"github.com/Saulorangel87/App-de-treino/backend/internal/xlsx"
)

type fakeExporter struct {
	userID string
	sheets []xlsx.Sheet
	err    error
}

func (f *fakeExporter) ExportAccountSpreadsheet(_ context.Context, userID string) ([]xlsx.Sheet, error) {
	f.userID = userID
	return f.sheets, f.err
}

func exportRequest(server *Server) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/account/export", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	response := httptest.NewRecorder()
	server.exportAccountData(response, req)
	return response
}

func sampleSheets() []xlsx.Sheet {
	return []xlsx.Sheet{
		{Name: "Conta", Header: []string{"Campo", "Valor"}, Rows: [][]any{{"E-mail", "a@example.invalid"}}},
		{Name: "Treinos realizados", Header: []string{"Data"}, Rows: [][]any{{"2026-09-30"}}},
	}
}

func TestExportAccountDataReturnsASpreadsheetForTheSignedInUser(t *testing.T) {
	store := httpTestAuthStore{user: auth.User{ID: "user-1"}}
	exporter := &fakeExporter{sheets: sampleSheets()}
	server := &Server{auth: auth.NewService(&store, time.Hour), exporter: exporter}

	response := exportRequest(server)
	if response.Code != http.StatusOK {
		t.Fatalf("export returned %d, want 200: %s", response.Code, response.Body)
	}
	if exporter.userID != "user-1" {
		t.Fatalf("exported data for %q, want the signed-in user", exporter.userID)
	}
	if got := response.Header().Get("Content-Type"); got != spreadsheetContentType {
		t.Fatalf("content type %q", got)
	}
	disposition := response.Header().Get("Content-Disposition")
	if !strings.HasPrefix(disposition, `attachment; filename="cadencia-dados-`) || !strings.HasSuffix(disposition, `.xlsx"`) {
		t.Fatalf("unexpected content disposition %q", disposition)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("the export must not be cached")
	}

	archive, err := zip.NewReader(bytes.NewReader(response.Body.Bytes()), int64(response.Body.Len()))
	if err != nil {
		t.Fatalf("the response is not an xlsx file: %v", err)
	}
	var account string
	for _, file := range archive.File {
		if file.Name == "xl/worksheets/sheet1.xml" {
			handle, _ := file.Open()
			body, _ := io.ReadAll(handle)
			handle.Close()
			account = string(body)
		}
	}
	if !strings.Contains(account, "a@example.invalid") || !strings.Contains(account, "Observação") {
		t.Fatalf("the account sheet is missing the e-mail or the note about the complete copy:\n%s", account)
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
