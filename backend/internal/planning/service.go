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
	StartWorkoutByUserID(context.Context, string, string) error
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
