package planning

import "time"

const localDateLayout = "2006-01-02"

// NormalizeLocalDate returns the athlete's current date (YYYY-MM-DD) to use when a
// workout is started. The client knows the athlete's local date, which the server
// (in UTC) does not: late in the evening in Brazil the UTC date is already
// tomorrow. The client's value is accepted only when it is a valid date within one
// day of the server's UTC date, which covers every time zone; anything else falls
// back to the UTC date, so a forged or stale value cannot unlock a future workout.
func NormalizeLocalDate(raw string, now time.Time) string {
	utcToday := now.UTC().Truncate(24 * time.Hour)
	fallback := utcToday.Format(localDateLayout)
	parsed, err := time.Parse(localDateLayout, raw)
	if err != nil || parsed.Format(localDateLayout) != raw {
		return fallback
	}
	if diff := parsed.Sub(utcToday); diff < -24*time.Hour || diff > 24*time.Hour {
		return fallback
	}
	return raw
}
