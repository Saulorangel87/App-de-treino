package httpapi

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type rateLimitEntry struct {
	count   int
	resetAt time.Time
}

type requestRateLimiter struct {
	mu         sync.Mutex
	entries    map[string]rateLimitEntry
	now        func() time.Time
	maxEntries int
}

func newRequestRateLimiter() *requestRateLimiter {
	return &requestRateLimiter{
		entries:    make(map[string]rateLimitEntry),
		now:        time.Now,
		maxEntries: 10_000,
	}
}

func (l *requestRateLimiter) limit(scope string, max int, window time.Duration, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowed, retryAfter := l.allow(scope+":"+clientAddress(r), max, window)
		if !allowed {
			seconds := int(retryAfter.Seconds())
			if seconds < 1 {
				seconds = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(seconds))
			writeError(w, http.StatusTooManyRequests, "rate_limited", "Muitas tentativas. Tente novamente mais tarde.")
			return
		}
		next(w, r)
	}
}

func (l *requestRateLimiter) allow(key string, max int, window time.Duration) (bool, time.Duration) {
	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()

	if entry, ok := l.entries[key]; ok && now.Before(entry.resetAt) {
		if entry.count >= max {
			return false, entry.resetAt.Sub(now)
		}
		entry.count++
		l.entries[key] = entry
		return true, 0
	}
	if len(l.entries) >= l.maxEntries {
		for entryKey, entry := range l.entries {
			if !now.Before(entry.resetAt) {
				delete(l.entries, entryKey)
			}
		}
		if len(l.entries) >= l.maxEntries {
			return false, time.Minute
		}
	}
	l.entries[key] = rateLimitEntry{count: 1, resetAt: now.Add(window)}
	return true, 0
}

func clientAddress(r *http.Request) string {
	if ip := net.ParseIP(strings.TrimSpace(r.Header.Get("CF-Connecting-IP"))); ip != nil {
		return ip.String()
	}
	if host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		if ip := net.ParseIP(host); ip != nil {
			return ip.String()
		}
	}
	if ip := net.ParseIP(strings.TrimSpace(r.RemoteAddr)); ip != nil {
		return ip.String()
	}
	return "unknown"
}

func csrfProtection(allowedOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if !requestOriginAllowed(r, allowedOrigin) {
			writeError(w, http.StatusForbidden, "csrf_origin_rejected", "A origem desta solicitação não é permitida.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requestOriginAllowed(r *http.Request, allowedOrigin string) bool {
	configured := canonicalOrigin(allowedOrigin)
	if configured == "" {
		return false
	}
	if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
		return canonicalOrigin(origin) == configured
	}
	if referer := strings.TrimSpace(r.Header.Get("Referer")); referer != "" {
		return canonicalOrigin(referer) == configured
	}
	_, hasSession := sessionCookie(r)
	return !hasSession
}

func canonicalOrigin(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil {
		return ""
	}
	return strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host)
}

func sessionCookie(r *http.Request) (*http.Cookie, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	return cookie, err == nil && cookie.Value != ""
}

func securityHeaders(secureCookies bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		if secureCookies {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		}
		next.ServeHTTP(w, r)
	})
}
