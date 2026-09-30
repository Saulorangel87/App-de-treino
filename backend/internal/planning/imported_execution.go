package planning

import (
	"math"
	"time"
)

const (
	importedExecutionVersion = "imported-execution-v1"
	importedExecutionMode    = "observation"
	importedExecutionScope   = "linked_imported_activity"
)

// ImportedActivityFacts are the measured values of an imported activity that the
// athlete linked to a workout. They come from the device file, not from what the
// athlete typed.
type ImportedActivityFacts struct {
	ActivityID        string
	Source            string
	StartedAt         time.Time
	MovingSeconds     int
	DistanceKM        *float64
	ElevationGainM    *int
	AverageHeartRate  *int
	MaxHeartRate      *int
	AveragePowerW     *int
	NormalizedPowerW  *int
	AverageCadenceRPM *int
}

// ImportedExecutionAssessment compares a workout's prescription with the activity
// measured by the athlete's device. It is observation only: it is attached when the
// plan is read, never stored, never read by buildPlan or by any adaptation, and it
// does not change the prescription.
type ImportedExecutionAssessment struct {
	Version                   string   `json:"version"`
	Mode                      string   `json:"mode"`
	Scope                     string   `json:"scope"`
	UsedForPrescription       bool     `json:"used_for_prescription"`
	ActivityID                string   `json:"activity_id"`
	Source                    string   `json:"source"`
	StartedAt                 string   `json:"started_at"`
	PlannedDurationMinutes    int      `json:"planned_duration_minutes"`
	MeasuredMovingMinutes     int      `json:"measured_moving_minutes"`
	DurationDeltaMinutes      int      `json:"duration_delta_minutes"`
	DurationCompletionPercent *float64 `json:"duration_completion_percent,omitempty"`
	DistanceKM                *float64 `json:"distance_km,omitempty"`
	ElevationGainM            *int     `json:"elevation_gain_m,omitempty"`
	AverageHeartRate          *int     `json:"average_heart_rate,omitempty"`
	MaxHeartRate              *int     `json:"max_heart_rate,omitempty"`
	AveragePowerW             *int     `json:"average_power_watts,omitempty"`
	NormalizedPowerW          *int     `json:"normalized_power_watts,omitempty"`
	AverageCadenceRPM         *int     `json:"average_cadence_rpm,omitempty"`
	// RecordedDurationMinutes is what the completed session registered (app timer
	// or correction), so the measured and recorded durations can be compared.
	RecordedDurationMinutes *int     `json:"recorded_duration_minutes,omitempty"`
	RecordedVsMeasuredDelta *int     `json:"recorded_vs_measured_delta_minutes,omitempty"`
	MissingData             []string `json:"missing_data"`
}

// AssessImportedExecution builds the observation for one workout and its linked
// imported activity. session may be nil when the workout was not completed in the app.
func AssessImportedExecution(workout Workout, facts ImportedActivityFacts) ImportedExecutionAssessment {
	measured := int(math.Round(float64(facts.MovingSeconds) / 60))
	result := ImportedExecutionAssessment{
		Version: importedExecutionVersion, Mode: importedExecutionMode, Scope: importedExecutionScope,
		UsedForPrescription: false,
		ActivityID:          facts.ActivityID, Source: facts.Source,
		StartedAt:              facts.StartedAt.UTC().Format(time.RFC3339),
		PlannedDurationMinutes: workout.DurationMinutes,
		MeasuredMovingMinutes:  measured,
		DurationDeltaMinutes:   measured - workout.DurationMinutes,
		DistanceKM:             facts.DistanceKM, ElevationGainM: facts.ElevationGainM,
		AverageHeartRate: facts.AverageHeartRate, MaxHeartRate: facts.MaxHeartRate,
		AveragePowerW: facts.AveragePowerW, NormalizedPowerW: facts.NormalizedPowerW,
		AverageCadenceRPM: facts.AverageCadenceRPM,
		MissingData:       []string{},
	}
	if workout.DurationMinutes > 0 {
		percent := math.Round(float64(measured)/float64(workout.DurationMinutes)*1000) / 10
		result.DurationCompletionPercent = &percent
	}
	if workout.Session != nil && workout.Session.Status == "completed" && workout.Session.DurationMinutes != nil {
		recorded := *workout.Session.DurationMinutes
		delta := recorded - measured
		result.RecordedDurationMinutes = &recorded
		result.RecordedVsMeasuredDelta = &delta
	} else {
		result.MissingData = append(result.MissingData, "completed_session")
	}
	if facts.AverageHeartRate == nil {
		result.MissingData = append(result.MissingData, "average_heart_rate")
	}
	if facts.AveragePowerW == nil {
		result.MissingData = append(result.MissingData, "average_power_watts")
	}
	if facts.AverageCadenceRPM == nil {
		result.MissingData = append(result.MissingData, "average_cadence_rpm")
	}
	return result
}
