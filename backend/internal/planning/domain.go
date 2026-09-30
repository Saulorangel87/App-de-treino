package planning

import (
	"errors"
	"regexp"
	"time"
)

var (
	ErrIncompleteOnboarding = errors.New("incomplete onboarding")
	ErrInvalidPlanID        = errors.New("invalid plan id")
	ErrPlanMissing          = errors.New("plan not found")
	ErrInvalidWorkoutID     = errors.New("invalid workout id")
	ErrWorkoutMissing       = errors.New("workout not found")
	ErrInvalidTransition    = errors.New("invalid workout transition")
	ErrWorkoutNotPast       = errors.New("workout date has not passed")
	ErrInvalidFeedback      = errors.New("invalid workout feedback")
	ErrInvalidCorrection    = errors.New("invalid workout correction")
	ErrWorkoutCorrection    = errors.New("workout correction not allowed")
	ErrWorkoutSafetyBlocked = errors.New("workout blocked by active safety limitation")
)

var planIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

type LimitationContext struct {
	Kind                             string
	ProfessionalClearanceRecommended bool
	MedicalRestriction               bool
	RecentSurgery                    bool
	ExerciseProhibited               bool
	ConditionAffectingExercise       bool
}

type AvailabilitySlot struct {
	Weekday          int
	AvailableMinutes int
	PreferredTime    *string
	Location         *string
}

type ProfileContext struct {
	BirthDate      *string
	Sex            *string
	HeightCM       *float64
	WeightKG       *float64
	WaistCM        *float64
	BodyFatPercent *float64
	WeightTrend    string
	ActivityLevel  *string
}

type CyclingContext struct {
	WeeklyHours            float64  `json:"weekly_hours"`
	PracticeDurationMonths int      `json:"practice_duration_months"`
	AverageRideMinutes     int      `json:"average_ride_minutes"`
	LongestRideMinutes     int      `json:"longest_ride_minutes"`
	WeeklyRides            int      `json:"weekly_rides"`
	RecentWeeklyDistanceKM float64  `json:"recent_weekly_distance_km"`
	RecentTrainingWeeks    int      `json:"recent_training_weeks"`
	TrainingStatus         string   `json:"training_status"`
	RecentBestDistanceKM   float64  `json:"recent_best_distance_km"`
	PreferredSessionTypes  []string `json:"preferred_session_types"`
	Discipline             string   `json:"discipline"`
	BikeType               string   `json:"bike_type"`
	Terrain                string   `json:"terrain"`
	UsesHeartRate          bool     `json:"uses_heart_rate"`
	UsesPower              bool     `json:"uses_power"`
	UsesGPS                bool     `json:"uses_gps"`
	UsesSportsWatch        bool     `json:"uses_sports_watch"`
	UsesSmartTrainer       bool     `json:"uses_smart_trainer"`
	FTP                    *int     `json:"ftp,omitempty"`
	FTPTestDate            *string  `json:"ftp_test_date,omitempty"`
	FTPProtocol            string   `json:"ftp_protocol,omitempty"`
	AveragePowerWatts      *int     `json:"average_power_watts,omitempty"`
	EventGoal              bool     `json:"event_goal"`
	EventDistanceKM        *int     `json:"event_distance_km,omitempty"`
	EventDate              *string  `json:"event_date,omitempty"`
}

// ObservedTrainingSummary is an aggregate of recent completed sessions and
// recovery check-ins. The motor uses it only as a conservative readiness
// signal, never as a diagnosis or a replacement for the athlete profile.
type ObservedTrainingSummary struct {
	WindowDays             int                   `json:"window_days"`
	CompletedSessions      int                   `json:"completed_sessions"`
	CompletedMinutes       int                   `json:"completed_minutes"`
	AverageRPE             float64               `json:"average_rpe"`
	AverageFatigue         float64               `json:"average_fatigue"`
	PainReported           bool                  `json:"pain_reported"`
	RecoveryCheckins       int                   `json:"recovery_checkins"`
	AverageRecoveryFatigue float64               `json:"average_recovery_fatigue"`
	DataCoverage           *ObservedDataCoverage `json:"data_coverage,omitempty"`
}

func (summary ObservedTrainingSummary) HasData() bool {
	return summary.CompletedSessions > 0 || summary.RecoveryCheckins > 0
}

func (summary ObservedTrainingSummary) RequiresRecovery() bool {
	return summary.PainReported || summary.AverageFatigue >= 4 || summary.AverageRecoveryFatigue >= 4
}

type Context struct {
	ProfileID              string
	Profile                ProfileContext
	ExperienceLevel        string
	PrimaryGoal            string
	SecondaryGoal          string
	Limitations            []LimitationContext
	Availability           []AvailabilitySlot
	Cycling                CyclingContext
	Observed               ObservedTrainingSummary
	TrainingHistory        []TrainingHistoryWindow
	TrainingHistoryPeriods []TrainingHistoryPeriod
	BaselineEligible       bool
	RotationIndex          int
	// RecentSignals are dated records of the last 14 days, used only to assess
	// Protection. Protection stays nil until protection levels are enabled, and
	// prescription then keeps the legacy Observed.RequiresRecovery() behavior.
	RecentSignals []RecentSignal
	Protection    *ProtectionAssessment
}

type Workout struct {
	ID              string          `json:"id,omitempty"`
	ScheduledOn     string          `json:"scheduled_on"`
	Name            string          `json:"name"`
	Objective       string          `json:"objective"`
	DurationMinutes int             `json:"duration_minutes"`
	TargetRPE       float64         `json:"target_rpe"`
	Structure       map[string]any  `json:"structure"`
	Explanation     map[string]any  `json:"explanation"`
	Status          string          `json:"status"`
	Session         *WorkoutSession `json:"session,omitempty"`
}

// WorkoutStep is the actionable sequence shown to the athlete. Keeping the
// steps in the prescription (rather than only in the interface) lets every
// client explain the same validated workout structure.
type WorkoutStep struct {
	Order           int     `json:"order"`
	Kind            string  `json:"kind"`
	Title           string  `json:"title"`
	DurationMinutes int     `json:"duration_minutes"`
	TargetRPE       float64 `json:"target_rpe"`
	Instruction     string  `json:"instruction"`
}

type WorkoutSession struct {
	ID                string     `json:"id"`
	Status            string     `json:"status"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	CancelledAt       *time.Time `json:"cancelled_at,omitempty"`
	DurationMinutes   *int       `json:"duration_minutes,omitempty"`
	ActualRPE         *float64   `json:"actual_rpe,omitempty"`
	DistanceKM        *float64   `json:"distance_km,omitempty"`
	ElevationGainM    *int       `json:"elevation_gain_m,omitempty"`
	AveragePowerW     *int       `json:"average_power_watts,omitempty"`
	AverageHeartRate  *int       `json:"average_heart_rate,omitempty"`
	AverageCadenceRPM *int       `json:"average_cadence_rpm,omitempty"`
	Feedback          *Feedback  `json:"feedback,omitempty"`
}

type Feedback struct {
	CompletionStatus   string `json:"completion_status"`
	PartialReason      string `json:"partial_reason,omitempty"`
	Difficulty         string `json:"difficulty"`
	PainReported       bool   `json:"pain_reported"`
	FatigueAfter       int    `json:"fatigue_after"`
	RecoveryAfter      *int   `json:"recovery_after,omitempty"`
	RepeatConfidence   *int   `json:"repeat_confidence,omitempty"`
	Satisfaction       *int   `json:"satisfaction,omitempty"`
	Terrain            string `json:"terrain,omitempty"`
	ExternalConditions string `json:"external_conditions,omitempty"`
	EquipmentUsed      string `json:"equipment_used,omitempty"`
	Notes              string `json:"notes,omitempty"`
}

type CompletionInput struct {
	CompletionStatus   string
	PartialReason      string
	ActualRPE          float64
	Difficulty         string
	PainReported       bool
	FatigueAfter       int
	RecoveryAfter      *int
	RepeatConfidence   *int
	Satisfaction       *int
	Terrain            string
	ExternalConditions string
	EquipmentUsed      string
	Notes              string
	DistanceKM         *float64
	ElevationGainM     *int
	AveragePowerW      *int
	AverageHeartRate   *int
	AverageCadenceRPM  *int
}

// WorkoutCorrectionInput replaces only optional pedal metrics on a completed
// session. Nil values intentionally clear the corresponding metric.
type WorkoutCorrectionInput struct {
	DistanceKM        *float64
	ElevationGainM    *int
	AveragePowerW     *int
	AverageHeartRate  *int
	AverageCadenceRPM *int
}

type Activity struct {
	ID                string     `json:"id"`
	WorkoutID         string     `json:"workout_id"`
	Name              string     `json:"name"`
	Objective         string     `json:"objective"`
	ScheduledOn       string     `json:"scheduled_on"`
	Status            string     `json:"status"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	CancelledAt       *time.Time `json:"cancelled_at,omitempty"`
	DurationMinutes   *int       `json:"duration_minutes,omitempty"`
	ActualRPE         *float64   `json:"actual_rpe,omitempty"`
	DistanceKM        *float64   `json:"distance_km,omitempty"`
	ElevationGainM    *int       `json:"elevation_gain_m,omitempty"`
	AveragePowerW     *int       `json:"average_power_watts,omitempty"`
	AverageHeartRate  *int       `json:"average_heart_rate,omitempty"`
	AverageCadenceRPM *int       `json:"average_cadence_rpm,omitempty"`
	Feedback          *Feedback  `json:"feedback,omitempty"`
}

type ScientificSource struct {
	SourceKey         string   `json:"source_key"`
	Title             string   `json:"title"`
	Authors           string   `json:"authors"`
	PublishedYear     int      `json:"published_year"`
	URL               string   `json:"url"`
	TrainingFocus     string   `json:"training_focus"`
	EvidenceLevel     string   `json:"evidence_level"`
	Summary           string   `json:"summary"`
	PopulationStudied string   `json:"population_studied"`
	ResearchObjective string   `json:"research_objective"`
	StimulusAnalyzed  string   `json:"stimulus_analyzed"`
	ExpectedBenefits  string   `json:"expected_benefits"`
	Limitations       string   `json:"limitations"`
	Risks             string   `json:"risks"`
	Contraindications string   `json:"contraindications"`
	ConfidenceLevel   string   `json:"confidence_level"`
	LastReviewedOn    string   `json:"last_reviewed_on"`
	RelatedRules      []string `json:"related_rules"`
}

type Plan struct {
	ID                   string             `json:"id,omitempty"`
	StartsOn             string             `json:"starts_on"`
	EndsOn               string             `json:"ends_on"`
	Status               string             `json:"status"`
	PrescriptionSnapshot map[string]any     `json:"prescription_snapshot"`
	Workouts             []Workout          `json:"workouts"`
	Evidence             []ScientificSource `json:"evidence,omitempty"`
}
