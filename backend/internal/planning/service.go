package planning

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
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
)

var planIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

type LimitationContext struct {
	Kind                             string
	ProfessionalClearanceRecommended bool
}

type AvailabilitySlot struct {
	Weekday          int
	AvailableMinutes int
	Location         *string
}

type CyclingContext struct {
	WeeklyHours            float64  `json:"weekly_hours"`
	LongestRideMinutes     int      `json:"longest_ride_minutes"`
	WeeklyRides            int      `json:"weekly_rides"`
	RecentWeeklyDistanceKM float64  `json:"recent_weekly_distance_km"`
	RecentTrainingWeeks    int      `json:"recent_training_weeks"`
	RecentBestDistanceKM   float64  `json:"recent_best_distance_km"`
	PreferredSessionTypes  []string `json:"preferred_session_types"`
	Discipline             string   `json:"discipline"`
	BikeType               string   `json:"bike_type"`
	Terrain                string   `json:"terrain"`
	UsesHeartRate          bool     `json:"uses_heart_rate"`
	UsesPower              bool     `json:"uses_power"`
	FTP                    *int     `json:"ftp,omitempty"`
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
	ExperienceLevel        string
	PrimaryGoal            string
	Limitations            []LimitationContext
	Availability           []AvailabilitySlot
	Cycling                CyclingContext
	Observed               ObservedTrainingSummary
	TrainingHistory        []TrainingHistoryWindow
	TrainingHistoryPeriods []TrainingHistoryPeriod
	BaselineEligible       bool
	RotationIndex          int
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
	ID               string     `json:"id"`
	Status           string     `json:"status"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	CancelledAt      *time.Time `json:"cancelled_at,omitempty"`
	DurationMinutes  *int       `json:"duration_minutes,omitempty"`
	ActualRPE        *float64   `json:"actual_rpe,omitempty"`
	DistanceKM       *float64   `json:"distance_km,omitempty"`
	ElevationGainM   *int       `json:"elevation_gain_m,omitempty"`
	AveragePowerW    *int       `json:"average_power_watts,omitempty"`
	AverageHeartRate *int       `json:"average_heart_rate,omitempty"`
	Feedback         *Feedback  `json:"feedback,omitempty"`
}

type Feedback struct {
	CompletionStatus string `json:"completion_status"`
	PartialReason    string `json:"partial_reason,omitempty"`
	Difficulty       string `json:"difficulty"`
	PainReported     bool   `json:"pain_reported"`
	FatigueAfter     int    `json:"fatigue_after"`
	RecoveryAfter    *int   `json:"recovery_after,omitempty"`
	RepeatConfidence *int   `json:"repeat_confidence,omitempty"`
	Notes            string `json:"notes,omitempty"`
}

type CompletionInput struct {
	CompletionStatus string
	PartialReason    string
	ActualRPE        float64
	Difficulty       string
	PainReported     bool
	FatigueAfter     int
	RecoveryAfter    *int
	RepeatConfidence *int
	Notes            string
	DistanceKM       *float64
	ElevationGainM   *int
	AveragePowerW    *int
	AverageHeartRate *int
}

// WorkoutCorrectionInput replaces only optional pedal metrics on a completed
// session. Nil values intentionally clear the corresponding metric.
type WorkoutCorrectionInput struct {
	DistanceKM       *float64
	ElevationGainM   *int
	AveragePowerW    *int
	AverageHeartRate *int
}

type Activity struct {
	ID               string     `json:"id"`
	WorkoutID        string     `json:"workout_id"`
	Name             string     `json:"name"`
	Objective        string     `json:"objective"`
	ScheduledOn      string     `json:"scheduled_on"`
	Status           string     `json:"status"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	CancelledAt      *time.Time `json:"cancelled_at,omitempty"`
	DurationMinutes  *int       `json:"duration_minutes,omitempty"`
	ActualRPE        *float64   `json:"actual_rpe,omitempty"`
	DistanceKM       *float64   `json:"distance_km,omitempty"`
	ElevationGainM   *int       `json:"elevation_gain_m,omitempty"`
	AveragePowerW    *int       `json:"average_power_watts,omitempty"`
	AverageHeartRate *int       `json:"average_heart_rate,omitempty"`
	Feedback         *Feedback  `json:"feedback,omitempty"`
}

type ScientificSource struct {
	SourceKey     string `json:"source_key"`
	Title         string `json:"title"`
	Authors       string `json:"authors"`
	PublishedYear int    `json:"published_year"`
	URL           string `json:"url"`
	TrainingFocus string `json:"training_focus"`
	EvidenceLevel string `json:"evidence_level"`
	Summary       string `json:"summary"`
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

type Store interface {
	PlanningContextByUserID(context.Context, string) (Context, error)
	SaveDraftPlan(context.Context, string, Plan) (Plan, error)
	CurrentPlanByUserID(context.Context, string) (Plan, error)
	ActivatePlanByUserID(context.Context, string, string) error
	StartWorkoutByUserID(context.Context, string, string) error
	CompleteWorkoutByUserID(context.Context, string, string, CompletionInput) error
	CorrectWorkoutDataByUserID(context.Context, string, string, WorkoutCorrectionInput) error
	CancelWorkoutByUserID(context.Context, string, string) error
	MarkWorkoutMissedByUserID(context.Context, string, string) error
	ActivitiesByUserID(context.Context, string) ([]Activity, error)
}

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service {
	return &Service{store: store, now: time.Now}
}

func (s *Service) Generate(ctx context.Context, userID string) (Plan, error) {
	input, err := s.store.PlanningContextByUserID(ctx, userID)
	if err != nil {
		return Plan{}, err
	}
	plan, err := buildPlan(input, s.now())
	if err != nil {
		return Plan{}, err
	}
	return s.store.SaveDraftPlan(ctx, input.ProfileID, plan)
}

func (s *Service) Current(ctx context.Context, userID string) (Plan, error) {
	return s.store.CurrentPlanByUserID(ctx, userID)
}

func (s *Service) Workout(ctx context.Context, userID, workoutID string) (Workout, error) {
	if !planIDPattern.MatchString(workoutID) {
		return Workout{}, ErrInvalidWorkoutID
	}
	plan, err := s.store.CurrentPlanByUserID(ctx, userID)
	if err != nil {
		return Workout{}, err
	}
	for _, workout := range plan.Workouts {
		if workout.ID == workoutID {
			return workout, nil
		}
	}
	return Workout{}, ErrWorkoutMissing
}

func (s *Service) Activate(ctx context.Context, userID, planID string) (Plan, error) {
	if !planIDPattern.MatchString(planID) {
		return Plan{}, ErrInvalidPlanID
	}
	if err := s.store.ActivatePlanByUserID(ctx, userID, planID); err != nil {
		return Plan{}, err
	}
	return s.store.CurrentPlanByUserID(ctx, userID)
}

func (s *Service) StartWorkout(ctx context.Context, userID, workoutID string) (Plan, error) {
	if !planIDPattern.MatchString(workoutID) {
		return Plan{}, ErrInvalidWorkoutID
	}
	if err := s.store.StartWorkoutByUserID(ctx, userID, workoutID); err != nil {
		return Plan{}, err
	}
	return s.store.CurrentPlanByUserID(ctx, userID)
}

func (s *Service) CompleteWorkout(ctx context.Context, userID, workoutID string, input CompletionInput) (Plan, error) {
	if !planIDPattern.MatchString(workoutID) {
		return Plan{}, ErrInvalidWorkoutID
	}
	if !validCompletion(input) {
		return Plan{}, ErrInvalidFeedback
	}
	if err := s.store.CompleteWorkoutByUserID(ctx, userID, workoutID, input); err != nil {
		return Plan{}, err
	}
	return s.store.CurrentPlanByUserID(ctx, userID)
}

func (s *Service) CorrectWorkout(ctx context.Context, userID, workoutID string, input WorkoutCorrectionInput) (Plan, error) {
	if !planIDPattern.MatchString(workoutID) {
		return Plan{}, ErrInvalidWorkoutID
	}
	if !validWorkoutCorrection(input) {
		return Plan{}, ErrInvalidCorrection
	}
	if err := s.store.CorrectWorkoutDataByUserID(ctx, userID, workoutID, input); err != nil {
		return Plan{}, err
	}
	return s.store.CurrentPlanByUserID(ctx, userID)
}

func (s *Service) CancelWorkout(ctx context.Context, userID, workoutID string) (Plan, error) {
	if !planIDPattern.MatchString(workoutID) {
		return Plan{}, ErrInvalidWorkoutID
	}
	if err := s.store.CancelWorkoutByUserID(ctx, userID, workoutID); err != nil {
		return Plan{}, err
	}
	return s.store.CurrentPlanByUserID(ctx, userID)
}

func (s *Service) MarkWorkoutMissed(ctx context.Context, userID, workoutID string) (Plan, error) {
	if !planIDPattern.MatchString(workoutID) {
		return Plan{}, ErrInvalidWorkoutID
	}
	if err := s.store.MarkWorkoutMissedByUserID(ctx, userID, workoutID); err != nil {
		return Plan{}, err
	}
	return s.store.CurrentPlanByUserID(ctx, userID)
}

func (s *Service) Activities(ctx context.Context, userID string) ([]Activity, error) {
	return s.store.ActivitiesByUserID(ctx, userID)
}

func validCompletion(input CompletionInput) bool {
	if input.ActualRPE < 1 || input.ActualRPE > 10 || input.FatigueAfter < 1 || input.FatigueAfter > 5 || len(input.Notes) > 1000 {
		return false
	}
	completionStatus := normalizeCompletionStatus(input.CompletionStatus)
	if completionStatus != "complete" && completionStatus != "partial" {
		return false
	}
	if completionStatus == "partial" && !validPartialReason(input.PartialReason) {
		return false
	}
	if completionStatus == "complete" && strings.TrimSpace(input.PartialReason) != "" {
		return false
	}
	if input.DistanceKM != nil && (*input.DistanceKM < 0 || *input.DistanceKM > 2000) {
		return false
	}
	if input.ElevationGainM != nil && (*input.ElevationGainM < 0 || *input.ElevationGainM > 20000) {
		return false
	}
	if input.AveragePowerW != nil && (*input.AveragePowerW < 0 || *input.AveragePowerW > 2000) {
		return false
	}
	if input.AverageHeartRate != nil && (*input.AverageHeartRate < 30 || *input.AverageHeartRate > 250) {
		return false
	}
	if input.RecoveryAfter != nil && (*input.RecoveryAfter < 1 || *input.RecoveryAfter > 5) {
		return false
	}
	if input.RepeatConfidence != nil && (*input.RepeatConfidence < 1 || *input.RepeatConfidence > 5) {
		return false
	}
	switch input.Difficulty {
	case "very_easy", "easy", "moderate", "hard", "very_hard":
		return true
	default:
		return false
	}
}

func validWorkoutCorrection(input WorkoutCorrectionInput) bool {
	if input.DistanceKM != nil && !finiteInRange(*input.DistanceKM, 0, 2000) {
		return false
	}
	if input.ElevationGainM != nil && (*input.ElevationGainM < 0 || *input.ElevationGainM > 20000) {
		return false
	}
	if input.AveragePowerW != nil && (*input.AveragePowerW < 0 || *input.AveragePowerW > 2000) {
		return false
	}
	if input.AverageHeartRate != nil && (*input.AverageHeartRate < 30 || *input.AverageHeartRate > 250) {
		return false
	}
	return true
}

func normalizeCompletionStatus(value string) string {
	if value == "" {
		return "complete"
	}
	return value
}

func validPartialReason(value string) bool {
	switch value {
	case "time_available_changed", "fatigue_or_recovery", "pain_or_discomfort", "equipment_or_conditions", "other":
		return true
	default:
		return false
	}
}

func buildPlan(input Context, now time.Time) (Plan, error) {
	if input.ProfileID == "" || input.PrimaryGoal == "" || len(input.Availability) == 0 {
		return Plan{}, ErrIncompleteOnboarding
	}
	maxSessions := map[string]int{"beginner": 3, "intermediate": 4, "advanced": 5}[input.ExperienceLevel]
	if maxSessions == 0 {
		return Plan{}, ErrIncompleteOnboarding
	}

	slots := append([]AvailabilitySlot(nil), input.Availability...)
	sort.Slice(slots, func(i, j int) bool { return slots[i].AvailableMinutes > slots[j].AvailableMinutes })
	if len(slots) > maxSessions {
		slots = slots[:maxSessions]
	}
	sort.Slice(slots, func(i, j int) bool { return slots[i].Weekday < slots[j].Weekday })

	start := nextMonday(now)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	restricted := len(input.Limitations) > 0
	for _, item := range input.Limitations {
		if item.ProfessionalClearanceRecommended {
			restricted = true
		}
	}
	workouts := make([]Workout, 0, len(slots)*4)
	multipliers := []float64{0.85, 0.95, 1.0, 0.75}
	eventTaper := assessEventTaper(input, now, restricted)
	for week := 0; week < 4; week++ {
		longIndex := longestSlot(slots)
		intensityIndex := intensitySlot(slots, longIndex)
		recoveryWeek := week == len(multipliers)-1
		for index, slot := range slots {
			scheduledOn := start.AddDate(0, 0, week*7+weekdayOffset(slot.Weekday))
			if week == 0 && scheduledOn.Before(today) {
				continue
			}
			kind := "base"
			if index == longIndex {
				kind = "long"
			} else if index == intensityIndex && !restricted && !recoveryWeek {
				kind = "quality"
			}
			workouts = append(workouts, makeWorkout(input, slot, kind, restricted, multipliers[week], week, scheduledOn, eventTaper))
		}
	}

	periodizationShadow := assessPeriodizationShadow(workouts, start, now)
	stimulusSelectionShadow := assessStimulusSelectionShadow(input, workouts, now, restricted)
	trainingHistory := buildTrainingHistorySnapshot(input.TrainingHistory, now, input.TrainingHistoryPeriods)
	planningCoherenceShadow := assessPlanningCoherenceShadow(periodizationShadow, trainingHistory.StimulusDistribution, stimulusSelectionShadow, now)
	plan := Plan{
		StartsOn: start.Format("2006-01-02"),
		EndsOn:   start.AddDate(0, 0, 27).Format("2006-01-02"),
		Status:   "draft",
		PrescriptionSnapshot: map[string]any{
			"engine_version":            "rules-v1",
			"event_taper":               eventTaper,
			"rules_v2_shadow":           assessRulesV2Shadow(input, now),
			"periodization_shadow":      periodizationShadow,
			"stimulus_selection_shadow": stimulusSelectionShadow,
			"planning_coherence_shadow": planningCoherenceShadow,
			"readiness_assessment":      assessReadiness(input, now),
			"training_history":          trainingHistory,
			"experience_level":          input.ExperienceLevel,
			"primary_goal":              input.PrimaryGoal,
			"restricted":                restricted,
			"sessions_per_week":         len(slots),
			"cycling_context": map[string]any{
				"weekly_hours":              input.Cycling.WeeklyHours,
				"longest_ride_minutes":      input.Cycling.LongestRideMinutes,
				"weekly_rides":              input.Cycling.WeeklyRides,
				"recent_weekly_distance_km": input.Cycling.RecentWeeklyDistanceKM,
				"recent_training_weeks":     input.Cycling.RecentTrainingWeeks,
				"recent_best_distance_km":   input.Cycling.RecentBestDistanceKM,
				"preferred_session_types":   input.Cycling.PreferredSessionTypes,
				"discipline":                input.Cycling.Discipline,
				"bike_type":                 input.Cycling.BikeType,
				"terrain":                   input.Cycling.Terrain,
				"uses_heart_rate":           input.Cycling.UsesHeartRate,
				"uses_power":                input.Cycling.UsesPower,
				"event_goal":                input.Cycling.EventGoal,
				"event_distance_km":         input.Cycling.EventDistanceKM,
				"event_date":                input.Cycling.EventDate,
			},
			"baseline_eligible": input.BaselineEligible,
			"rotation_index":    input.RotationIndex,
			"observed_training": map[string]any{
				"window_days":              input.Observed.WindowDays,
				"completed_sessions":       input.Observed.CompletedSessions,
				"completed_minutes":        input.Observed.CompletedMinutes,
				"average_rpe":              input.Observed.AverageRPE,
				"average_fatigue":          input.Observed.AverageFatigue,
				"pain_reported":            input.Observed.PainReported,
				"recovery_checkins":        input.Observed.RecoveryCheckins,
				"average_recovery_fatigue": input.Observed.AverageRecoveryFatigue,
				"requires_recovery":        input.Observed.RequiresRecovery(),
			},
		},
		Workouts: workouts,
	}
	return plan, nil
}

func makeWorkout(input Context, slot AvailabilitySlot, kind string, restricted bool, multiplier float64, weekIndex int, date time.Time, eventTaper EventTaperAssessment) Workout {
	baseMinutes := map[string]int{"beginner": 45, "intermediate": 60, "advanced": 75}[input.ExperienceLevel]
	name := "Giro de base"
	targetRPE := 4.0
	mainBlock := "Ritmo confortável e contínuo"
	summary := explanationFor(kind, restricted)
	usesControlledIntervals := false
	usesRoadModerateIntervals := false
	usesRoadHighIntensityIntervals := false
	usesRoadVO2Intervals := false
	usesShortSelfRegulatedIntervals := false
	usesXCOAerobicIntervals := false
	rotationApplied := false
	activeRecoveryApplied := false
	eventTaperApplied := false
	eventSpecificPhase := eventSpecificPhase(input.Cycling, date)
	observedProtected := input.Observed.RequiresRecovery() && (input.Observed.PainReported || kind == "quality")
	if kind == "base" && weekIndex == 3 {
		name = "Recuperação ativa"
		targetRPE = 3.5
		mainBlock = "Pedale leve e contínuo, sem transformar a sessão em treino de qualidade"
		summary = "A semana de recuperação reduz a carga e usa um giro ativo para manter o movimento sem acrescentar estímulo intenso."
		activeRecoveryApplied = true
	}
	if kind == "long" {
		baseMinutes = map[string]int{"beginner": 75, "intermediate": 90, "advanced": 120}[input.ExperienceLevel]
		name = "Endurance contínuo"
		targetRPE = 5.0
		mainBlock = "Volume aeróbico estável"
	}
	if kind == "quality" {
		name = "Tempo controlado"
		targetRPE = 6.0
		mainBlock = "3 blocos sustentados com recuperação leve"
		preference := preferredQualityPreference(input.Cycling)
		if input.Cycling.Discipline == "road" && input.ExperienceLevel == "advanced" && input.BaselineEligible && (input.PrimaryGoal == "performance" || input.PrimaryGoal == "event") && input.Cycling.RecentTrainingWeeks >= 8 && input.Cycling.WeeklyRides >= 3 && preference == "vo2max" && slot.AvailableMinutes >= 60 && multiplier >= 0.95 && (!input.Cycling.EventGoal || eventSpecificPhase) {
			name = "Intervalos VO₂max de estrada"
			targetRPE = 8.0
			mainBlock = "4 blocos de 4 min em esforço muito forte-controlado com 4 min leves entre os blocos"
			summary = "A modalidade de estrada, a preferência explícita, o objetivo, a avaliação apta e o histórico mínimo permitem um piloto de VO₂max conservador; a sessão não usa sprint máximo nem meta fixa de potência."
			usesRoadVO2Intervals = true
		} else if (input.Cycling.Discipline == "road" || input.Cycling.Discipline == "indoor") && input.ExperienceLevel == "advanced" && input.BaselineEligible && (input.PrimaryGoal == "performance" || input.PrimaryGoal == "event") && input.Cycling.RecentTrainingWeeks >= 8 && input.Cycling.WeeklyRides >= 3 && preference == "short_intervals" && slot.AvailableMinutes >= 50 && multiplier >= 0.95 && (!input.Cycling.EventGoal || eventSpecificPhase) {
			name = "Intervalos curtos autorregulados"
			targetRPE = 7.5
			mainBlock = "6 blocos de 1 min em RPE 7–8 com 1 min leve entre os blocos"
			summary = "A modalidade, a preferência explícita, o objetivo, a avaliação apta e o histórico mínimo permitem um piloto curto autorregulado; a sessão não usa sprint máximo, potência fixa ou cadência obrigatória."
			usesShortSelfRegulatedIntervals = true
		} else if input.Cycling.Discipline == "road" && input.ExperienceLevel != "beginner" && input.BaselineEligible && (input.PrimaryGoal == "performance" || input.PrimaryGoal == "event") && slot.AvailableMinutes >= 60 && multiplier >= 0.95 && (preference == "" || preference == "intervals") && (!input.Cycling.EventGoal || eventSpecificPhase) {
			if input.ExperienceLevel == "advanced" && input.Cycling.RecentTrainingWeeks >= 8 && input.Cycling.WeeklyRides >= 3 && input.RotationIndex%2 == 1 && slot.AvailableMinutes >= 75 {
				name = "Intervalos intensos de estrada"
				targetRPE = 8.0
				mainBlock = "5 blocos intensos de 8 min com 4 min leves entre os blocos"
				summary = "A modalidade de estrada, o objetivo, a avaliação apta e a rotação do ciclo permitem um piloto intenso restrito; a sessão não reproduz a carga dos estudos."
				usesRoadHighIntensityIntervals = true
			} else {
				name = "Intervalos moderados de estrada"
				targetRPE = 6.0
				mainBlock = "3 blocos moderados de 10 min com 3 min leves entre os blocos"
				summary = "A modalidade de estrada, o objetivo e a avaliação submáxima apta permitem um piloto intervalado moderado e conservador."
				usesRoadModerateIntervals = true
			}
		} else if input.Cycling.Discipline == "mtb_xco" && input.ExperienceLevel == "advanced" && input.BaselineEligible && (input.PrimaryGoal == "performance" || input.PrimaryGoal == "event") && slot.AvailableMinutes >= 75 && multiplier >= 0.95 && (preference == "" || preference == "intervals") && (!input.Cycling.EventGoal || eventSpecificPhase) {
			name = "Intervalos aeróbicos XCO"
			targetRPE = 7.0
			mainBlock = "5 blocos aeróbicos de 4 min com 4 min leves entre os blocos"
			summary = "A modalidade XCO explícita, o objetivo compatível e a avaliação submáxima apta liberam um piloto aeróbico conservador; o treino não simula trechos técnicos nem usa sprint máximo."
			usesXCOAerobicIntervals = true
		} else if input.ExperienceLevel == "advanced" && input.Cycling.EventGoal && input.BaselineEligible && eventSpecificPhase {
			name = "Ritmo de prova controlado"
			targetRPE = 6.5
			mainBlock = "3 blocos em ritmo sustentável, com recuperação leve"
			summary = "A prova está próxima o suficiente para orientar um estímulo sustentável, sem simular a prova inteira."
		} else if preference == "cadence" && input.ExperienceLevel != "beginner" {
			name = "Cadência técnica"
			targetRPE = 5.0
			mainBlock = "6 blocos de cadência controlada com recuperação leve"
			summary = "A preferência por cadência orienta uma sessão técnica com esforço controlado."
		} else if preference == "hills" && input.Cycling.Terrain == "hilly" && input.ExperienceLevel != "beginner" {
			name = "Subidas controladas"
			if input.ExperienceLevel == "advanced" {
				targetRPE = 6.5
			}
			mainBlock = "4 blocos sustentados em subida, com recuperação leve"
			summary = "A preferência por subidas e o terreno informado orientam um estímulo controlado."
		} else if preference == "intervals" && input.ExperienceLevel == "advanced" && input.BaselineEligible && (input.PrimaryGoal == "performance" || input.PrimaryGoal == "event") && slot.AvailableMinutes >= 50 && multiplier >= 0.95 {
			name = "Intervalos controlados"
			targetRPE = 7.0
			mainBlock = "4 blocos de 4 min em esforço forte-controlado, com 3 min leves entre os blocos"
			summary = "A preferência por intervalos foi combinada com uma avaliação submáxima apta e uma semana de construção."
			usesControlledIntervals = true
		} else if preference == "sweet_spot" && input.ExperienceLevel == "advanced" && input.Cycling.UsesPower && input.Cycling.FTP != nil {
			name = "Sweet spot por potência"
			targetRPE = 7.0
			mainBlock = "3 blocos sustentados guiados pelo FTP informado, com recuperação leve"
			summary = "A preferência por sweet spot foi combinada com o medidor de potência e o FTP informado."
		} else if input.ExperienceLevel == "advanced" && input.BaselineEligible && (input.PrimaryGoal == "performance" || input.PrimaryGoal == "event") && slot.AvailableMinutes >= 50 && multiplier >= 0.95 {
			name = "Intervalos controlados"
			targetRPE = 7.0
			mainBlock = "4 blocos de 4 min em esforço forte-controlado, com 3 min leves entre os blocos"
			summary = "A referência submáxima concluída sem sinais de alerta permite uma progressão intervalada controlada."
			usesControlledIntervals = true
		} else if input.ExperienceLevel == "intermediate" && input.Cycling.BikeType == "indoor" {
			name = "Cadência técnica"
			targetRPE = 5.0
			mainBlock = "6 blocos de cadência controlada com recuperação leve"
			summary = "A sessão usa o ambiente indoor informado para praticar cadência com esforço controlado."
		} else if input.Cycling.Terrain == "hilly" && input.ExperienceLevel != "beginner" {
			name = "Subidas controladas"
			if input.ExperienceLevel == "advanced" {
				targetRPE = 6.5
			}
			mainBlock = "4 blocos sustentados em subida, com recuperação leve"
			summary = "O terreno com subidas informado orienta um estímulo controlado e específico."
		} else if input.ExperienceLevel == "advanced" && input.Cycling.UsesPower && input.Cycling.FTP != nil {
			name = "Sweet spot por potência"
			targetRPE = 7.0
			mainBlock = "3 blocos sustentados guiados pelo FTP informado, com recuperação leve"
			summary = "O medidor de potência e o FTP informados permitem orientar um esforço sustentável."
		} else if input.ExperienceLevel == "advanced" {
			name = "Sweet spot progressivo"
			targetRPE = 7.0
		}
		if input.ExperienceLevel == "advanced" && input.RotationIndex%2 == 1 {
			switch name {
			case "Intervalos controlados":
				name = "Sweet spot progressivo"
				mainBlock = "3 blocos sustentados com recuperação leve"
				summary = "A rotação entre ciclos alterna o estímulo intervalado por um esforço sustentável de qualidade."
				usesControlledIntervals = false
				rotationApplied = true
			case "Subidas controladas", "Sweet spot por potência", "Sweet spot progressivo":
				name = "Tempo controlado"
				targetRPE = 6.0
				mainBlock = "3 blocos sustentados com recuperação leve"
				summary = "A rotação entre ciclos alterna o estímulo específico sem aumentar a carga planejada."
				rotationApplied = true
			}
		}
	}
	if restricted {
		rotationApplied = false
		activeRecoveryApplied = false
		name = "Giro leve protegido"
		targetRPE = 3.5
		mainBlock = "Esforço leve; interromper diante de dor ou desconforto"
		if baseMinutes > 45 {
			baseMinutes = 45
		}
		multiplier *= 0.8
		summary = explanationFor(kind, true)
	} else if observedProtected {
		rotationApplied = false
		activeRecoveryApplied = false
		name = "Giro leve protegido"
		targetRPE = 3.5
		mainBlock = "Esforço leve; interromper diante de dor ou desconforto"
		if baseMinutes > 45 {
			baseMinutes = 45
		}
		multiplier *= 0.8
		summary = "O histórico recente de esforço, fadiga ou dor recomenda uma sessão leve e protegida neste ciclo."
	}
	duration := int(float64(baseMinutes) * multiplier)
	if duration < 20 {
		duration = 20
	}
	if duration > slot.AvailableMinutes {
		duration = slot.AvailableMinutes
	}
	if !restricted && !observedProtected && eventTaperAppliesToWorkout(input.Cycling, eventTaper, date, weekIndex == 3) {
		eventTaperApplied = true
		duration = int(float64(duration) * eventTaper.VolumeMultiplier)
		if duration < 20 {
			duration = 20
		}
		summary += " O volume foi reduzido para a janela pré-prova, sem aumentar a intensidade nem a frequência planejada."
	}
	protocol := protocolForWorkout(name)

	rules := []string{
		fmt.Sprintf("Agendado em um dia com %d minutos disponíveis.", slot.AvailableMinutes),
		fmt.Sprintf("Carga compatível com experiência %s.", experienceLabel(input.ExperienceLevel)),
		"Progressão de três semanas seguida por uma semana de recuperação.",
	}
	if input.Cycling.WeeklyRides > 0 || input.Cycling.RecentWeeklyDistanceKM > 0 {
		rules = append(rules, "Histórico recente informado usado para contextualizar a sessão.")
	}
	if input.Observed.HasData() {
		rules = append(rules, fmt.Sprintf("Histórico observado dos últimos %d dias considerado (%d sessões concluídas).", input.Observed.WindowDays, input.Observed.CompletedSessions))
	}
	if preference := preferredQualityPreference(input.Cycling); kind == "quality" && preference != "" {
		rules = append(rules, fmt.Sprintf("Preferência por %s considerada dentro dos limites de segurança.", sessionPreferenceLabel(preference)))
	}
	if usesControlledIntervals {
		rules = append(rules, "Intervalos liberados pela avaliação submáxima apta, apenas nas semanas de construção.")
	}
	if usesRoadModerateIntervals {
		rules = append(rules, "Piloto de estrada moderado liberado por modalidade explícita, objetivo compatível, avaliação apta e disponibilidade suficiente.")
	}
	if usesRoadHighIntensityIntervals {
		rules = append(rules, "Piloto intenso de estrada liberado apenas para atleta avançado elegível, em ciclo alternado e semana de construção; sem reprodução da carga estudada.")
	}
	if usesRoadVO2Intervals {
		rules = append(rules, "Piloto de VO₂max de estrada liberado por preferência explícita, modalidade, objetivo, avaliação apta e histórico mínimo; sem sprint máximo ou meta fixa de potência.")
	}
	if usesShortSelfRegulatedIntervals {
		rules = append(rules, "Piloto de intervalos curtos liberado por preferência explícita, modalidade, objetivo, avaliação apta e histórico mínimo; esforço autorregulado, sem sprint máximo ou meta fixa de potência.")
	}
	if usesXCOAerobicIntervals {
		rules = append(rules, "Piloto aeróbico XCO liberado por modalidade explícita, objetivo compatível, avaliação apta e disponibilidade suficiente; sem sprint máximo ou simulação técnica.")
	}
	if rotationApplied {
		rules = append(rules, "Sessão alternada pela rotação explicável do ciclo, sem aumentar a carga planejada.")
	}
	if activeRecoveryApplied {
		rules = append(rules, "Variação de recuperação ativa aplicada na quarta semana, sem aumentar a carga planejada.")
	}
	if eventTaperApplied {
		rules = append(rules, "Taper pré-prova aplicado nesta sessão: volume reduzido de forma conservadora, mantendo a frequência planejada.")
	}
	if restricted {
		rules = append(rules, "Intensidade limitada por uma condição de segurança ativa.")
	}
	if observedProtected {
		rules = append(rules, "Sessão protegida por sinais recentes de recuperação insuficiente ou dor relatada.")
	}
	evidenceKeys := append([]string(nil), protocol.EvidenceKeys...)
	evidenceScope := protocol.EvidenceScope
	if eventTaperApplied {
		evidenceKeys = append(append([]string(nil), eventTaper.EvidenceKeys...), evidenceKeys...)
		evidenceScope += " O taper pré-prova usa evidência de redução de volume em ciclistas/endurance, com transferência limitada a atletas elegíveis; não é dose universal."
	}
	return Workout{
		ScheduledOn:     date.Format("2006-01-02"),
		Name:            name,
		Objective:       objectiveFor(input.PrimaryGoal),
		DurationMinutes: duration,
		TargetRPE:       targetRPE,
		Structure:       buildStructure(duration, targetRPE, name, mainBlock),
		Explanation:     map[string]any{"summary": summary, "rules": rules, "protocol_key": protocol.Key, "evidence_keys": evidenceKeys, "evidence_scope": evidenceScope, "event_taper_applied": eventTaperApplied},
		Status:          "planned",
	}
}

func preferredQualityPreference(context CyclingContext) string {
	if len(context.PreferredSessionTypes) == 0 || len(context.PreferredSessionTypes) >= 7 {
		return ""
	}
	for _, preference := range []string{"vo2max", "short_intervals", "intervals", "sweet_spot", "hills", "cadence"} {
		for _, selected := range context.PreferredSessionTypes {
			if selected == preference {
				return preference
			}
		}
	}
	return ""
}

func sessionPreferenceLabel(preference string) string {
	labels := map[string]string{"cadence": "cadência", "hills": "subidas", "intervals": "intervalos", "sweet_spot": "sweet spot", "vo2max": "VO₂max", "short_intervals": "intervalos curtos"}
	return labels[preference]
}

func buildStructure(duration int, targetRPE float64, name, mainBlock string) map[string]any {
	warmup := minInt(10, maxInt(5, duration/6))
	cooldown := minInt(10, maxInt(5, duration/8))
	mainMinutes := maxInt(1, duration-warmup-cooldown)
	steps := []WorkoutStep{
		{Order: 1, Kind: "warmup", Title: "Aquecimento", DurationMinutes: warmup, TargetRPE: 3, Instruction: "Pedale de forma confortável e aumente o ritmo aos poucos."},
	}

	protocol := protocolForWorkout(name)
	var mainSteps []WorkoutStep
	if protocol.Repetitions > 1 {
		mainSteps = repeatedSteps(mainMinutes, protocol.WorkMinutes, protocol.RecoveryMinutes, protocol.Repetitions, targetRPE, protocol.WorkTitle, protocol.WorkInstruction)
	} else {
		mainSteps = []WorkoutStep{{Kind: "main", Title: "Parte principal", DurationMinutes: mainMinutes, TargetRPE: targetRPE, Instruction: mainBlock + "."}}
	}

	for index := range mainSteps {
		mainSteps[index].Order = len(steps) + 1
		steps = append(steps, mainSteps[index])
	}
	steps = append(steps, WorkoutStep{
		Order: len(steps) + 1, Kind: "cooldown", Title: "Desaquecimento", DurationMinutes: cooldown,
		TargetRPE: 2.5, Instruction: "Reduza o ritmo gradualmente e termine pedalando leve.",
	})

	return map[string]any{
		"warmup_minutes":   warmup,
		"main":             mainBlock,
		"cooldown_minutes": cooldown,
		"protocol_key":     protocol.Key,
		"steps":            steps,
	}
}

func repeatedSteps(mainMinutes, workMinutes, recoveryMinutes, repetitions int, targetRPE float64, title, instruction string) []WorkoutStep {
	if mainMinutes <= 0 {
		return nil
	}
	for repetitions > 1 && repetitions*workMinutes+(repetitions-1)*recoveryMinutes > mainMinutes {
		repetitions--
	}
	for workMinutes > 3 && repetitions*workMinutes+(repetitions-1)*recoveryMinutes > mainMinutes {
		workMinutes--
	}
	if repetitions < 2 || repetitions*workMinutes+(repetitions-1)*recoveryMinutes > mainMinutes {
		return []WorkoutStep{{Kind: "main", Title: "Parte principal", DurationMinutes: mainMinutes, TargetRPE: targetRPE, Instruction: instruction}}
	}

	steps := make([]WorkoutStep, 0, repetitions*2+1)
	used := 0
	for index := 0; index < repetitions; index++ {
		steps = append(steps, WorkoutStep{Kind: "work", Title: fmt.Sprintf("%s %d de %d", title, index+1, repetitions), DurationMinutes: workMinutes, TargetRPE: targetRPE, Instruction: instruction})
		used += workMinutes
		if index < repetitions-1 {
			steps = append(steps, WorkoutStep{Kind: "recovery", Title: "Recuperação leve", DurationMinutes: recoveryMinutes, TargetRPE: 2.5, Instruction: "Pedale bem leve até recuperar a respiração antes do próximo bloco."})
			used += recoveryMinutes
		}
	}
	if remaining := mainMinutes - used; remaining > 0 {
		steps = append(steps, WorkoutStep{Kind: "easy", Title: "Pedal leve contínuo", DurationMinutes: remaining, TargetRPE: 3, Instruction: "Complete o tempo restante em ritmo confortável, sem forçar."})
	}
	return steps
}

func nextMonday(value time.Time) time.Time {
	local := time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
	if local.Weekday() == time.Sunday {
		return local.AddDate(0, 0, 1)
	}
	return local.AddDate(0, 0, -((int(local.Weekday()) + 6) % 7))
}

func weekdayOffset(weekday int) int { return (weekday + 6) % 7 }

func eventSpecificPhase(cycling CyclingContext, workoutDate time.Time) bool {
	daysUntilEvent, ok := daysUntilEvent(cycling, workoutDate)
	return ok && daysUntilEvent >= 0 && daysUntilEvent <= 42
}

func longestSlot(slots []AvailabilitySlot) int {
	longest := 0
	for index := range slots {
		if slots[index].AvailableMinutes > slots[longest].AvailableMinutes {
			longest = index
		}
	}
	return longest
}

func intensitySlot(slots []AvailabilitySlot, longIndex int) int {
	for index := range slots {
		if index != longIndex {
			return index
		}
	}
	return -1
}

func objectiveFor(goal string) string {
	labels := map[string]string{
		"health":            "Construir saúde cardiovascular com consistência",
		"fitness":           "Desenvolver condicionamento geral",
		"endurance":         "Aumentar resistência para pedais mais longos",
		"performance":       "Elevar a capacidade de sustentar esforço",
		"event":             "Criar base específica para o evento-alvo",
		"weight_management": "Apoiar gasto energético com carga sustentável",
	}
	return labels[goal]
}

func explanationFor(kind string, restricted bool) string {
	if restricted {
		return "Sessão deliberadamente leve para respeitar a limitação informada."
	}
	if kind == "long" {
		return "O maior período disponível da semana recebe o estímulo principal de resistência."
	}
	if kind == "quality" {
		return "Uma única sessão de qualidade oferece estímulo sem concentrar carga excessiva."
	}
	return "Sessão de base para acumular consistência com baixo custo de recuperação."
}

func experienceLabel(value string) string {
	return map[string]string{"beginner": "iniciante", "intermediate": "intermediária", "advanced": "avançada"}[value]
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
