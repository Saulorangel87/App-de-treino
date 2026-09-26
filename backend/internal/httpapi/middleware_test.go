package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testLogger(buffer *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buffer, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestObservabilityRecoversFromPanicWithJSONError(t *testing.T) {
	var logs bytes.Buffer
	handler := observability(testLogger(&logs), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/me", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}
	var body struct {
		Error struct{ Code string } `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Error.Code != "internal_error" {
		t.Fatalf("unexpected body %q (%v)", response.Body.String(), err)
	}
	if response.Header().Get(requestIDHeader) == "" {
		t.Fatal("expected a request id header")
	}
	if !strings.Contains(logs.String(), "request panic") {
		t.Fatalf("expected panic log, got %s", logs.String())
	}
}

func TestObservabilityLogsRequestsAndSkipsHealthyProbes(t *testing.T) {
	var logs bytes.Buffer
	handler := observability(testLogger(&logs), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))
	if logs.Len() != 0 {
		t.Fatalf("healthy probe should not be logged, got %s", logs.String())
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	request.Header.Set(requestIDHeader, "abcdef123456")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Header().Get(requestIDHeader) != "abcdef123456" {
		t.Fatalf("expected the caller request id to be kept, got %q", response.Header().Get(requestIDHeader))
	}
	if !strings.Contains(logs.String(), `"status":204`) || !strings.Contains(logs.String(), `"request_id":"abcdef123456"`) {
		t.Fatalf("unexpected access log %s", logs.String())
	}
}

func TestSanitizeRequestIDRejectsUnsafeValues(t *testing.T) {
	for _, value := range []string{"", "short", "has space in it", "line\nbreak-injection", strings.Repeat("a", 65)} {
		if sanitizeRequestID(value) != "" {
			t.Fatalf("%q should be rejected", value)
		}
	}
	if sanitizeRequestID("Abc-123_xyz9") == "" {
		t.Fatal("expected a safe id to be accepted")
	}
}
