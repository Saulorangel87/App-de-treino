package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

const requestIDHeader = "X-Request-ID"

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if !r.wroteHeader {
		r.status = status
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(body)
}

func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

// observability assigns a request id, recovers from panics with a JSON 500 and
// writes one structured access log line per request. Health probes are only
// logged when they fail, to keep the log readable.
func observability(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		requestID := sanitizeRequestID(r.Header.Get(requestIDHeader))
		if requestID == "" {
			requestID = newRequestID()
		}
		w.Header().Set(requestIDHeader, requestID)
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("request panic", "request_id", requestID, "method", r.Method, "path", r.URL.Path, "panic", recovered, "stack", string(debug.Stack()))
				if !recorder.wroteHeader {
					writeError(recorder, http.StatusInternalServerError, "internal_error", "Ocorreu um erro inesperado. Tente novamente.")
				}
			}
			route := r.Pattern
			if route == "" {
				route = "unmatched"
			}
			isProbe := r.URL.Path == "/health" || r.URL.Path == "/ready"
			if isProbe && recorder.status < 500 {
				return
			}
			level := slog.LevelInfo
			if recorder.status >= 500 {
				level = slog.LevelError
			} else if recorder.status >= 400 {
				level = slog.LevelWarn
			}
			logger.Log(r.Context(), level, "http request",
				"request_id", requestID, "method", r.Method, "route", route,
				"status", recorder.status, "duration_ms", time.Since(started).Milliseconds(),
				"client_ip", clientAddress(r))
		}()
		next.ServeHTTP(recorder, r)
	})
}

func newRequestID() string {
	raw := make([]byte, 8)
	if _, err := rand.Read(raw); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(raw)
}

// sanitizeRequestID accepts a caller-provided id only when it is short and made
// of safe characters, so it cannot inject log content or headers.
func sanitizeRequestID(value string) string {
	if len(value) < 8 || len(value) > 64 {
		return ""
	}
	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z', char >= 'A' && char <= 'Z', char >= '0' && char <= '9', char == '-', char == '_':
		default:
			return ""
		}
	}
	return value
}
