package auth

import "time"

// LoginFailures is the count of wrong passwords kept for one key (an IP
// address or an e-mail, both hashed). The HTTP layer decides the rules; the
// stores only keep the state.
type LoginFailures struct {
	Failures      int
	LastFailureAt time.Time
	BlockedUntil  time.Time
}
