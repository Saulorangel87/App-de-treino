package planning

import (
	"context"
	"time"
)

type Store interface {
	PlanningContextByUserID(context.Context, string) (Context, error)
	SaveDraftPlan(context.Context, string, Plan) (Plan, error)
	CurrentPlanByUserID(context.Context, string) (Plan, error)
	ActivatePlanByUserID(context.Context, string, string) error
	// StartWorkoutByUserID starts the session; today is the athlete's local date
	// (YYYY-MM-DD) and a workout scheduled after it cannot be started.
	StartWorkoutByUserID(ctx context.Context, userID, workoutID, today string) error
	// LogWorkoutByUserID records a planned workout as done with the duration the
	// athlete reported (no stopwatch). today is the athlete's local date.
	LogWorkoutByUserID(ctx context.Context, userID, workoutID string, input LogWorkoutInput, today string) error
	// UndoWorkoutByUserID erases a completed, skipped or in-progress record and
	// returns the workout to the plan, reverting the adaptations it caused.
	UndoWorkoutByUserID(ctx context.Context, userID, workoutID string) error
	CompleteWorkoutByUserID(context.Context, string, string, CompletionInput) error
	CorrectWorkoutDataByUserID(context.Context, string, string, WorkoutCorrectionInput) error
	CancelWorkoutByUserID(context.Context, string, string) error
	MarkWorkoutMissedByUserID(context.Context, string, string) error
	ActivitiesByUserID(context.Context, string) ([]Activity, error)
	// PlannedWorkoutsForReevaluation returns the still-planned workouts of the
	// active plan scheduled between the two dates (inclusive, YYYY-MM-DD).
	PlannedWorkoutsForReevaluation(ctx context.Context, userID, from, to string) ([]Workout, error)
	// ApplyWorkoutRevisions rewrites workouts that are still planned in the
	// active plan; anything else is left untouched. It returns how many changed.
	ApplyWorkoutRevisions(ctx context.Context, userID string, revisions []WorkoutRevision) (int, error)
	// RecordRecoverySelfReport stores "I am recovered" for today. Repeating it on
	// the same day is a no-op.
	RecordRecoverySelfReport(ctx context.Context, userID string) error
}

// WorkoutRevision is the rebuilt content of a planned workout.
type WorkoutRevision struct {
	WorkoutID       string
	Name            string
	Objective       string
	DurationMinutes int
	TargetRPE       float64
	Structure       map[string]any
	Explanation     map[string]any
}

type Service struct {
	store            Store
	now              func() time.Time
	protectionLevels bool
}

type Option func(*Service)

// WithProtectionLevels switches prescription and re-evaluation to the graduated,
// dated protection levels. Without it the legacy 28-day rule applies unchanged.
func WithProtectionLevels(enabled bool) Option {
	return func(s *Service) { s.protectionLevels = enabled }
}

func NewService(store Store, options ...Option) *Service {
	service := &Service{store: store, now: time.Now}
	for _, option := range options {
		option(service)
	}
	return service
}

// withProtection attaches the dated protection assessment when enabled.
func (s *Service) withProtection(input Context) Context {
	if s.protectionLevels {
		assessment := assessProtection(input.RecentSignals, s.now())
		input.Protection = &assessment
	}
	return input
}

func (s *Service) Generate(ctx context.Context, userID string) (Plan, error) {
	input, err := s.store.PlanningContextByUserID(ctx, userID)
	if err != nil {
		return Plan{}, err
	}
	input = s.withProtection(input)
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

// StartWorkout starts a workout scheduled for today or earlier. localDate is the
// athlete's date as the client sees it; see NormalizeLocalDate.
func (s *Service) StartWorkout(ctx context.Context, userID, workoutID, localDate string) (Plan, error) {
	if !planIDPattern.MatchString(workoutID) {
		return Plan{}, ErrInvalidWorkoutID
	}
	if err := s.store.StartWorkoutByUserID(ctx, userID, workoutID, NormalizeLocalDate(localDate, s.now())); err != nil {
		return Plan{}, err
	}
	return s.store.CurrentPlanByUserID(ctx, userID)
}

// LogWorkout marks a planned workout as done ("task mode"): the athlete reports
// the duration and the day instead of using the stopwatch. The session is kept
// with the duration source "reported", so calibration can tell it from a measured
// one. The workout must be scheduled for today or earlier, and the ride cannot be
// dated before it or more than MaxLogBackdateDays ago.
func (s *Service) LogWorkout(ctx context.Context, userID, workoutID string, input LogWorkoutInput, localDate string) (Plan, error) {
	if !planIDPattern.MatchString(workoutID) {
		return Plan{}, ErrInvalidWorkoutID
	}
	if !validCompletion(input.CompletionInput) {
		return Plan{}, ErrInvalidFeedback
	}
	if input.DurationMinutes < 1 || input.DurationMinutes > MaxLoggedDurationMinutes {
		return Plan{}, ErrInvalidLog
	}
	today := NormalizeLocalDate(localDate, s.now())
	if input.PerformedOn == "" {
		input.PerformedOn = today
	}
	performed, err := time.Parse(localDateLayout, input.PerformedOn)
	if err != nil || performed.Format(localDateLayout) != input.PerformedOn {
		return Plan{}, ErrInvalidLog
	}
	todayDate, _ := time.Parse(localDateLayout, today)
	if performed.After(todayDate) || performed.Before(todayDate.AddDate(0, 0, -MaxLogBackdateDays)) {
		return Plan{}, ErrInvalidLog
	}
	if err := s.store.LogWorkoutByUserID(ctx, userID, workoutID, input, today); err != nil {
		return Plan{}, err
	}
	s.reevaluateBestEffort(ctx, userID)
	return s.store.CurrentPlanByUserID(ctx, userID)
}

// UndoWorkout erases the record of a workout and gives it back to the plan, so a
// test or mistaken session does not stay in the history. Protection is
// reevaluated because the session's pain and fatigue no longer count.
func (s *Service) UndoWorkout(ctx context.Context, userID, workoutID string) (Plan, error) {
	if !planIDPattern.MatchString(workoutID) {
		return Plan{}, ErrInvalidWorkoutID
	}
	if err := s.store.UndoWorkoutByUserID(ctx, userID, workoutID); err != nil {
		return Plan{}, err
	}
	s.reevaluateBestEffort(ctx, userID)
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
	s.reevaluateBestEffort(ctx, userID)
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
	s.reevaluateBestEffort(ctx, userID)
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
