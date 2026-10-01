package planning

import (
	"testing"
	"time"
)

func TestNormalizeLocalDate(t *testing.T) {
	// 01:30 UTC on Oct 2 is still the evening of Oct 1 in São Paulo.
	now := time.Date(2026, 10, 2, 1, 30, 0, 0, time.UTC)
	cases := map[string]string{
		"2026-10-01": "2026-10-01", // athlete's evening, one day behind UTC
		"2026-10-02": "2026-10-02", // same as UTC
		"2026-10-03": "2026-10-03", // ahead of UTC (e.g. Asia/Oceania)
		"2026-10-04": "2026-10-02", // two days ahead: forged
		"2026-09-29": "2026-10-02", // stale
		"":           "2026-10-02",
		"amanhã":     "2026-10-02",
		"2026-13-45": "2026-10-02",
		"2026-10-1":  "2026-10-02", // not the exact format
	}
	for raw, want := range cases {
		if got := NormalizeLocalDate(raw, now); got != want {
			t.Errorf("NormalizeLocalDate(%q) = %s, want %s", raw, got, want)
		}
	}
}
