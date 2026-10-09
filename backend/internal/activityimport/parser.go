// Package activityimport reads athlete-uploaded .fit and .gpx files and
// extracts a per-activity summary (duration, distance, elevation, heart
// rate, power, cadence). It never reads a full second-by-second stream and
// it never influences rules-v1's prescription: imported data is only ever
// written to a workout_session through the existing completion/correction
// endpoints, after the athlete reviews and confirms it.
//
// See docs/proxima-fase-dados-reais.md for the product scope and the
// decisions behind it (Strava adiado; only .fit/.gpx in this version).
package activityimport

import (
	"errors"
	"time"
)

// SportCycling is the only sport accepted in this version; the whole product
// is scoped to cycling (see docs/roadmap-acceptance.md, seção 14).
const SportCycling = "cycling"

var (
	// ErrUnsupportedSport is returned when the file clearly describes a
	// non-cycling activity (e.g. a run exported by the same device).
	ErrUnsupportedSport = errors.New("activityimport: o arquivo não é de uma atividade de ciclismo")
	// ErrNoData is returned when the file has no session summary and no
	// records to fall back to (an empty or non-activity file).
	ErrNoData = errors.New("activityimport: o arquivo não tem dados de atividade")
	// ErrCorruptFile is returned when the file cannot be parsed at all.
	ErrCorruptFile = errors.New("activityimport: não foi possível ler o arquivo")
)

// Parsed is the subset of an activity file this package understands, already
// converted to the physical units used across the app (seconds, kilometers,
// meters, bpm, watts, rpm) — the same units workout_sessions already stores.
type Parsed struct {
	StartedAt         time.Time `json:"started_at"`
	MovingSeconds     int       `json:"moving_seconds"`
	DistanceKM        float64   `json:"distance_km"`
	ElevationGainM    *int      `json:"elevation_gain_m,omitempty"`
	AverageHeartRate  *int      `json:"average_heart_rate,omitempty"`
	MaxHeartRate      *int      `json:"max_heart_rate,omitempty"`
	AveragePowerW     *int      `json:"average_power_watts,omitempty"`
	NormalizedPowerW  *int      `json:"normalized_power_watts,omitempty"`
	AverageCadenceRPM *int      `json:"average_cadence_rpm,omitempty"`
	// Best20MinPowerW é a maior potência média em 20 minutos seguidos, calculada
	// dos registros segundo a segundo. Serve só para sugerir o FTP ao atleta.
	Best20MinPowerW *int `json:"best_20min_power_watts,omitempty"`
	// LocalDateKnown is true when StartedAt's calendar date reflects the
	// athlete's local time (the .fit file carried the device's timezone
	// offset). Files are stored in UTC; without this, a ride close to
	// midnight local time can land on the wrong UTC calendar day and miss
	// the planned workout it belongs to. GPX has no timezone field at all,
	// so this is always false for it; Service.Import widens the search by a
	// day in either direction when false, instead of matching exactly.
	LocalDateKnown bool `json:"-"`
}

// avg returns the mean of the samples that are non-nil, or nil when there is
// no sample at all — matching how the app already treats optional metrics.
func avgInt(samples []int) *int {
	if len(samples) == 0 {
		return nil
	}
	sum := 0
	for _, v := range samples {
		sum += v
	}
	mean := sum / len(samples)
	return &mean
}
