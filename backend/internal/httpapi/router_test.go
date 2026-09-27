package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/email"
)

type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

func TestHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	NewRouter(fakePinger{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, email.DevelopmentSender{}, "http://localhost:3000", "http://localhost:3000", false, true, 7*24*time.Hour, 24*time.Hour).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestReadyReturnsUnavailableWhenDatabaseFails(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/ready", nil)
	response := httptest.NewRecorder()
	NewRouter(fakePinger{err: errors.New("offline")}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, email.DevelopmentSender{}, "http://localhost:3000", "http://localhost:3000", false, true, 7*24*time.Hour, 24*time.Hour).ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", response.Code)
	}
}

type schemaPinger struct {
	fakePinger
	schemaErr error
}

func (p schemaPinger) Check(context.Context) error { return p.schemaErr }

func TestReadyReportsSchemaBehind(t *testing.T) {
	newRouter := func(pinger Pinger) http.Handler {
		return NewRouter(pinger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, email.DevelopmentSender{}, "http://localhost:3000", "http://localhost:3000", false, true, 7*24*time.Hour, 24*time.Hour)
	}
	behind := httptest.NewRecorder()
	newRouter(schemaPinger{schemaErr: errors.New("missing 000030")}).ServeHTTP(behind, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if behind.Code != http.StatusServiceUnavailable || !strings.Contains(behind.Body.String(), "schema_behind") {
		t.Fatalf("expected 503 schema_behind, got %d %s", behind.Code, behind.Body.String())
	}
	if strings.Contains(behind.Body.String(), "000030") {
		t.Fatal("the public readiness response must not leak migration names")
	}
	ok := httptest.NewRecorder()
	newRouter(schemaPinger{}).ServeHTTP(ok, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if ok.Code != http.StatusOK {
		t.Fatalf("expected 200 when the schema is current, got %d", ok.Code)
	}
}
