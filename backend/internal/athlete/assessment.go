package athlete

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidAssessment = errors.New("invalid assessment data")

type Assessment struct {
	ID                     string     `json:"id"`
	AssessmentType         string     `json:"assessment_type"`
	CompletedAt            *time.Time `json:"completed_at,omitempty"`
	DurationMinutes        int        `json:"duration_minutes"`
	TargetRPE              float64    `json:"target_rpe"`
	ActualRPE              float64    `json:"actual_rpe"`
	PainReported           bool       `json:"pain_reported"`
	Notes                  string     `json:"notes,omitempty"`
	EligibleForProgression bool       `json:"eligible_for_progression"`
}

type AssessmentStore interface {
	CurrentAssessmentByUserID(context.Context, string) (*Assessment, error)
	SaveSubmaxAssessment(context.Context, string, Assessment) (Assessment, error)
}

type AssessmentService struct{ store AssessmentStore }

func NewAssessmentService(store AssessmentStore) *AssessmentService {
	return &AssessmentService{store: store}
}

func (s *AssessmentService) Current(ctx context.Context, userID string) (*Assessment, error) {
	return s.store.CurrentAssessmentByUserID(ctx, userID)
}

func (s *AssessmentService) SaveSubmax(ctx context.Context, userID string, input Assessment) (Assessment, error) {
	if input.DurationMinutes < 15 || input.DurationMinutes > 30 || input.ActualRPE < 1 || input.ActualRPE > 10 || len(input.Notes) > 1000 {
		return Assessment{}, ErrInvalidAssessment
	}
	input.AssessmentType = "submax_reference"
	input.TargetRPE = 5
	input.EligibleForProgression = !input.PainReported && input.DurationMinutes >= 18 && input.ActualRPE <= 6
	return s.store.SaveSubmaxAssessment(ctx, userID, input)
}
