package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
)

// Wrong passwords at sign-in, per IP: the 4th answers with a warning that one
// attempt is left, the 5th blocks that IP for 10 minutes. Errors more than 10
// minutes apart do not add up. A correct login does not clear the IP count,
// otherwise someone with an account of their own could interleave successful
// logins and guess without limit.
//
// Per account (e-mail) there is never a block, so nobody can lock the owner out
// by typing wrong passwords for them. Against guesses spread over many IPs, the
// answer for that e-mail gets slower with each error (up to 5 seconds), and the
// count clears when the owner signs in.
const (
	maxLoginFailuresPerIP      = 5
	loginFailureWindow         = 10 * time.Minute
	loginBlockDuration         = 10 * time.Minute
	accountFailureWindow       = time.Hour
	accountSlowdownAfter       = 5
	accountSlowdownStep        = 500 * time.Millisecond
	accountSlowdownMax         = 5 * time.Second
	memoryLoginFailuresMaxKeys = 10_000
)

// LoginAttemptStore keeps the wrong-password counts. Production uses the
// database (repository.Store), so a restart or deploy does not reset them.
type LoginAttemptStore interface {
	LoginFailures(ctx context.Context, key string) (auth.LoginFailures, error)
	UpdateLoginFailures(ctx context.Context, key string, update func(auth.LoginFailures) auth.LoginFailures) (auth.LoginFailures, error)
	ClearLoginFailures(ctx context.Context, key string) error
}

func WithLoginAttempts(store LoginAttemptStore) RouterOption {
	return func(s *Server) { s.loginAttempts = store }
}

// loginKey hashes the IP or e-mail so neither is stored in clear text.
func loginKey(kind, value string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(value))))
	return kind + ":" + hex.EncodeToString(sum[:])
}

// countFailure adds one wrong password. With max > 0, reaching it blocks the
// key for block and starts a new count for when the block ends.
func countFailure(state auth.LoginFailures, now time.Time, max int, window, block time.Duration) auth.LoginFailures {
	if state.Failures > 0 && now.Sub(state.LastFailureAt) > window {
		state.Failures = 0
	}
	if !state.BlockedUntil.IsZero() && !now.Before(state.BlockedUntil) {
		state.BlockedUntil = time.Time{}
	}
	state.Failures++
	state.LastFailureAt = now
	if max > 0 && state.Failures >= max {
		state.BlockedUntil = now.Add(block)
		state.Failures = 0
	}
	return state
}

// accountSlowdown is the wait before checking a password for an e-mail that
// has had many recent errors, from any IP.
func accountSlowdown(state auth.LoginFailures, now time.Time) time.Duration {
	if state.Failures <= accountSlowdownAfter || now.Sub(state.LastFailureAt) > accountFailureWindow {
		return 0
	}
	return min(accountSlowdownMax, time.Duration(state.Failures-accountSlowdownAfter)*accountSlowdownStep)
}

func waitFor(ctx context.Context, delay time.Duration) {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

// loginFailuresOf reads a key; a store error is logged and treated as no
// errors, so a database hiccup never locks everyone out.
func (s *Server) loginFailuresOf(ctx context.Context, key string) auth.LoginFailures {
	state, err := s.loginAttempts.LoginFailures(ctx, key)
	if err != nil {
		slog.Default().Error("reading login failures failed", "error", err)
		return auth.LoginFailures{}
	}
	return state
}

// memoryLoginAttempts is the store used without a database (tests and tools).
// When full, it drops the oldest key instead of refusing new ones.
type memoryLoginAttempts struct {
	mu      sync.Mutex
	entries map[string]auth.LoginFailures
	maxKeys int
}

func newMemoryLoginAttempts() *memoryLoginAttempts {
	return &memoryLoginAttempts{entries: make(map[string]auth.LoginFailures), maxKeys: memoryLoginFailuresMaxKeys}
}

func (m *memoryLoginAttempts) LoginFailures(_ context.Context, key string) (auth.LoginFailures, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.entries[key], nil
}

func (m *memoryLoginAttempts) UpdateLoginFailures(_ context.Context, key string, update func(auth.LoginFailures) auth.LoginFailures) (auth.LoginFailures, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.entries[key]; !ok && len(m.entries) >= m.maxKeys {
		oldestKey, oldest := "", time.Time{}
		for entryKey, entry := range m.entries {
			if oldestKey == "" || entry.LastFailureAt.Before(oldest) {
				oldestKey, oldest = entryKey, entry.LastFailureAt
			}
		}
		delete(m.entries, oldestKey)
	}
	state := update(m.entries[key])
	m.entries[key] = state
	return state, nil
}

func (m *memoryLoginAttempts) ClearLoginFailures(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.entries, key)
	return nil
}
