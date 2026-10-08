package httpapi

import (
	"archive/zip"
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"github.com/Saulorangel87/App-de-treino/backend/internal/i18n"
	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
	"github.com/Saulorangel87/App-de-treino/backend/internal/xlsx"
	"golang.org/x/crypto/bcrypt"
)

func serveWithLanguage(handler http.HandlerFunc, request *http.Request) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	withLanguage(observability(slog.New(slog.NewTextHandler(io.Discard, nil)), handler)).ServeHTTP(response, request)
	return response
}

func TestErrorsFollowTheLanguageTheAppAsksFor(t *testing.T) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, "workout_not_found", "Treino não encontrado.")
	}
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept-Language", "en")
	english := serveWithLanguage(handler, request)
	if !strings.Contains(english.Body.String(), `"message":"Workout not found."`) || !strings.Contains(english.Body.String(), `"code":"workout_not_found"`) {
		t.Fatalf("English error: %s", english.Body)
	}
	if !strings.Contains(english.Header().Get("Vary"), "Accept-Language") {
		t.Fatalf("responses must vary by language, got Vary %q", english.Header().Get("Vary"))
	}

	portuguese := serveWithLanguage(handler, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(portuguese.Body.String(), "Treino não encontrado.") {
		t.Fatalf("a request without a language must stay in Portuguese: %s", portuguese.Body)
	}
}

func TestPanicMessageIsTranslatedToo(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept-Language", "en")
	response := serveWithLanguage(func(http.ResponseWriter, *http.Request) { panic("boom") }, request)
	if response.Code != http.StatusInternalServerError || !strings.Contains(response.Body.String(), "An unexpected error occurred") {
		t.Fatalf("panic response: %d %s", response.Code, response.Body)
	}
}

func TestWorkoutsAreTranslatedOnTheWayOutAndNotInTheEngine(t *testing.T) {
	workout := planning.Workout{
		Name: "Giro de base", Objective: "Desenvolver condicionamento geral", Status: "planned",
		Structure:   map[string]any{"main": "Ritmo confortável e contínuo", "steps": []planning.WorkoutStep{{Kind: "warmup", Title: "Aquecimento"}}},
		Explanation: map[string]any{"rules": []string{"Carga compatível com experiência iniciante."}, "protocol_key": "base_endurance"},
	}
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	response := serveWithLanguage(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"workout": workout, "duration": 75})
	}, request)
	body := response.Body.String()
	for _, want := range []string{`"name":"Base ride"`, `"objective":"Develop general fitness"`, `"main":"Comfortable, continuous pace"`, `"title":"Warm-up"`, `"Load suited to beginner experience."`, `"protocol_key":"base_endurance"`, `"status":"planned"`, `"duration":75`} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %s in %s", want, body)
		}
	}
	if workout.Name != "Giro de base" {
		t.Fatal("translating the response must not change the engine's value")
	}
}

func TestAccountDeletionAcceptsTheEnglishPhrase(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("uma-senha-segura"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	store := &accountDeletionStore{httpTestAuthStore: httpTestAuthStore{user: auth.User{ID: "user-1", PasswordHash: string(hash)}}}
	server := &Server{auth: auth.NewService(store, time.Hour)}
	request := httptest.NewRequest(http.MethodDelete, "/v1/auth/account", strings.NewReader(`{"password":"uma-senha-segura","confirmation":"delete account"}`))
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	response := httptest.NewRecorder()
	server.deleteAccount(response, request)
	if response.Code != http.StatusNoContent || !store.deleted {
		t.Fatalf("English confirmation returned %d with deleted=%t", response.Code, store.deleted)
	}
}

func TestEmailsAreWrittenInTheLanguageOfTheRequest(t *testing.T) {
	english := passwordResetMessage(i18n.English, "a@example.invalid", "Ana <b>", "https://app.example/redefinir-senha?token=x&y=1")
	if english.Subject != "Reset your password on Cadência" || !strings.Contains(english.HTML, "Hi, Ana &lt;b&gt;.") || !strings.Contains(english.HTML, "token=x&amp;y=1") {
		t.Fatalf("English reset email: %+v", english)
	}
	portuguese := verificationMessage(i18n.Portuguese, "a@example.invalid", "Ana", "https://app.example/verificar-email?token=x")
	if portuguese.Subject != "Confirme seu e-mail no Cadência" || !strings.HasPrefix(portuguese.Text, "Confirme seu e-mail: ") {
		t.Fatalf("Portuguese verification email: %+v", portuguese)
	}
	if verificationMessage(i18n.English, "a", "Ana", "u").Subject != "Confirm your email on Cadência" {
		t.Fatal("English verification email subject")
	}
}

func TestEnglishExportTranslatesLabelsWithoutChangingThePortugueseOnes(t *testing.T) {
	header := []string{"Data planejada", "Treino"}
	sheets := func() []xlsx.Sheet {
		return []xlsx.Sheet{
			{Name: "Conta", Header: []string{"Campo", "Valor"}, Rows: [][]any{{"E-mail confirmado", "Sim"}}},
			{Name: "Treinos realizados", Header: header, Rows: [][]any{{"30/09/2026", "Pedal longo"}}},
		}
	}
	store := httpTestAuthStore{user: auth.User{ID: "user-1"}}
	server := &Server{auth: auth.NewService(&store, time.Hour), exporter: &fakeExporter{sheets: sheets()}}
	request := httptest.NewRequest(http.MethodGet, "/v1/auth/account/export", nil)
	request.Header.Set("Accept-Language", "en")
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	response := serveWithLanguage(server.exportAccountData, request)
	if response.Code != http.StatusOK {
		t.Fatalf("export returned %d: %s", response.Code, response.Body)
	}
	if !strings.Contains(response.Header().Get("Content-Disposition"), "cadencia-data-") {
		t.Fatalf("English file name: %q", response.Header().Get("Content-Disposition"))
	}
	text := spreadsheetText(t, response.Body.Bytes())
	for _, want := range []string{"Account", "Field", "Email confirmed", "Yes", "Completed workouts", "Planned date", "Long ride", "This spreadsheet has the essentials"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in the English spreadsheet", want)
		}
	}
	if header[0] != "Data planejada" {
		t.Fatal("the shared Portuguese header must not be changed by an English export")
	}
}

// spreadsheetText returns the text stored in the workbook (shared strings and
// sheet names), enough to check which labels it carries.
func spreadsheetText(t *testing.T, body []byte) string {
	t.Helper()
	archive, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatalf("open spreadsheet: %v", err)
	}
	var text strings.Builder
	for _, file := range archive.File {
		if !strings.HasSuffix(file.Name, ".xml") {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatalf("open %s: %v", file.Name, err)
		}
		content, _ := io.ReadAll(reader)
		reader.Close()
		text.Write(content)
	}
	return text.String()
}
