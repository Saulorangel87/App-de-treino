package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

func TestHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	NewRouter(fakePinger{}, nil, nil, nil, nil, nil, nil, nil, "http://localhost:3000", false, 7*24*time.Hour).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestReadyReturnsUnavailableWhenDatabaseFails(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/ready", nil)
	response := httptest.NewRecorder()
	NewRouter(fakePinger{err: errors.New("offline")}, nil, nil, nil, nil, nil, nil, nil, "http://localhost:3000", false, 7*24*time.Hour).ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", response.Code)
	}
}
