package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"github.com/jackc/pgx/v5"
)

type fakeSource struct {
	users    map[string]auth.User
	sections map[string]json.RawMessage
	err      error
	exported []string
}

func (f *fakeSource) UserByEmail(_ context.Context, email string) (auth.User, error) {
	user, ok := f.users[email]
	if !ok {
		return auth.User{}, pgx.ErrNoRows
	}
	return user, nil
}

func (f *fakeSource) ExportAccountData(_ context.Context, userID string) (map[string]json.RawMessage, error) {
	f.exported = append(f.exported, userID)
	return f.sections, f.err
}

func newFake() *fakeSource {
	return &fakeSource{
		users:    map[string]auth.User{"atleta@exemplo.com": {ID: "user-1", Email: "atleta@exemplo.com"}},
		sections: map[string]json.RawMessage{"account": json.RawMessage(`{"email":"atleta@exemplo.com"}`), "limitations": json.RawMessage(`[{"description":"joelho"}]`)},
	}
}

func runWith(store source, args ...string) (string, string, error) {
	var out, logs bytes.Buffer
	err := run(context.Background(), args, store, &out, slog.New(slog.NewJSONHandler(&logs, nil)), func() time.Time {
		return time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	})
	return out.String(), logs.String(), err
}

func TestRunWritesTheCompleteCopyAndLogsOnlyTheTrace(t *testing.T) {
	store := newFake()
	out, logs, err := runWith(store, "--email", "  Atleta@Exemplo.com ")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(store.exported) != 1 || store.exported[0] != "user-1" {
		t.Fatalf("exported %v, want user-1 only", store.exported)
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal([]byte(out), &document); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out)
	}
	for _, key := range []string{"format_version", "exported_at", "notice", "account", "limitations"} {
		if _, ok := document[key]; !ok {
			t.Errorf("missing %q", key)
		}
	}
	if !strings.Contains(out, "joelho") {
		t.Error("the copy must include health data")
	}
	if !strings.Contains(logs, "atleta@exemplo.com") || strings.Contains(logs, "joelho") {
		t.Errorf("the log must name the account but never carry its content:\n%s", logs)
	}
}

func TestRunRejectsBadInputWithoutTouchingTheDatabase(t *testing.T) {
	for name, args := range map[string][]string{
		"sem e-mail":      {},
		"e-mail vazio":    {"--email", " "},
		"e-mail inválido": {"--email", "isso nao e email"},
		"com nome":        {"--email", "Fulano <atleta@exemplo.com>"},
		"flag inválida":   {"--cpf", "1"},
	} {
		store := newFake()
		out, _, err := runWith(store, args...)
		if err == nil || out != "" || len(store.exported) != 0 {
			t.Errorf("%s: err=%v out=%q exported=%v, want an error and no export", name, err, out, store.exported)
		}
	}
}

func TestRunReportsAnUnknownAccountWithoutOutput(t *testing.T) {
	store := newFake()
	out, _, err := runWith(store, "--email", "ninguem@exemplo.com")
	if err == nil || !strings.Contains(err.Error(), "no account") || out != "" {
		t.Fatalf("err=%v out=%q, want a clear error and no output", err, out)
	}
}

func TestRunFailsWhenTheExportFails(t *testing.T) {
	store := newFake()
	store.err = errors.New("db down")
	out, _, err := runWith(store, "--email", "atleta@exemplo.com")
	if err == nil || out != "" {
		t.Fatalf("err=%v out=%q, want an error and no partial output", err, out)
	}
}

func TestRunPropagatesWriteErrors(t *testing.T) {
	err := run(context.Background(), []string{"--email", "atleta@exemplo.com"}, newFake(), failingWriter{}, slog.New(slog.NewJSONHandler(io.Discard, nil)), time.Now)
	if err == nil {
		t.Fatal("a failed write must be reported")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }
