package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRequestRateLimiterBlocksAndExpires(t *testing.T) {
	current := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	limiter := &requestRateLimiter{
		entries:    make(map[string]rateLimitEntry),
		now:        func() time.Time { return current },
		maxEntries: 10,
	}
	handler := limiter.limit("login", 2, time.Minute, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	for attempt := 0; attempt < 2; attempt++ {
		request := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
		request.RemoteAddr = "192.0.2.10:1234"
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusNoContent {
			t.Fatalf("attempt %d returned %d, want 204", attempt+1, response.Code)
		}
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
	request.RemoteAddr = "192.0.2.10:1234"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusTooManyRequests || response.Header().Get("Retry-After") != "60" {
		t.Fatalf("blocked attempt returned %d with Retry-After %q", response.Code, response.Header().Get("Retry-After"))
	}

	current = current.Add(time.Minute)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("attempt after window returned %d, want 204", response.Code)
	}
}

func TestRequestRateLimiterPrefersCloudflareClientIP(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.RemoteAddr = "192.0.2.10:1234"
	request.Header.Set("CF-Connecting-IP", "198.51.100.20")
	if got := clientAddress(request); got != "198.51.100.20" {
		t.Fatalf("client address = %q, want Cloudflare address", got)
	}
}

func TestCSRFProtectionRejectsForeignOriginWithSession(t *testing.T) {
	handler := csrfProtection("https://cadencia.example", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodPost, "/v1/profile", nil)
	request.Header.Set("Origin", "https://evil.example")
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("foreign origin returned %d, want 403", response.Code)
	}
}

func TestCSRFProtectionAllowsConfiguredOrigin(t *testing.T) {
	handler := csrfProtection("https://cadencia.example", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodPost, "/v1/profile", nil)
	request.Header.Set("Origin", "https://cadencia.example")
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("configured origin returned %d, want 204", response.Code)
	}
}

func TestSecurityHeadersAreSetOnlyForSecureDeployment(t *testing.T) {
	handler := securityHeaders(true, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	for name, want := range map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Permissions-Policy":        "camera=(), microphone=(), geolocation=()",
		"Strict-Transport-Security": "max-age=31536000",
	} {
		if got := response.Header().Get(name); got != want {
			t.Fatalf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestDecodeJSONRejectsUnknownAndOversizedBodies(t *testing.T) {
	t.Run("unknown field", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"known":"ok","unexpected":true}`))
		response := httptest.NewRecorder()
		var input struct {
			Known string `json:"known"`
		}
		if decodeJSON(response, request, &input) {
			t.Fatal("unknown field was accepted")
		}
		if response.Code != http.StatusBadRequest {
			t.Fatalf("unknown field returned %d, want 400", response.Code)
		}
	})

	t.Run("oversized body", func(t *testing.T) {
		body := `{"known":"` + strings.Repeat("a", 1<<20) + `"}`
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		response := httptest.NewRecorder()
		var input struct {
			Known string `json:"known"`
		}
		if decodeJSON(response, request, &input) {
			t.Fatal("oversized body was accepted")
		}
		if response.Code != http.StatusBadRequest {
			t.Fatalf("oversized body returned %d, want 400", response.Code)
		}
	})
}
