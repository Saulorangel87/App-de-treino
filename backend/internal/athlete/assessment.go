package athlete

import (
	"context"
	"errors"
	"math"
	"time"
)

var ErrInvalidAssessment = errors.New("invalid assessment data")

const (
	// EfficiencyPowerPerBPM is watts per beat per minute, used when the athlete
	// has a power meter.
	EfficiencyPowerPerBPM = "power_per_bpm"
	// EfficiencySpeedPer100BPM is km/h for each 100 bpm, used when there is only
	// distance and heart rate.
	EfficiencySpeedPer100BPM = "speed_per_100bpm"

	// maxPlausibleSpeedKMH rejects a distance that does not fit the duration.
	maxPlausibleSpeedKMH = 80.0
	historyLimit         = 10
)

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

	// Optional numbers from the ride. They never change EligibleForProgression;
	// they only feed the efficiency and the drift below.
	AverageHeartRate    *int     `json:"average_heart_rate,omitempty"`
	AveragePowerW       *int     `json:"average_power_w,omitempty"`
	DistanceKM          *float64 `json:"distance_km,omitempty"`
	HeartRateFirstHalf  *int     `json:"heart_rate_first_half,omitempty"`
	HeartRateSecondHalf *int     `json:"heart_rate_second_half,omitempty"`

	// Derived on read, never stored.
	HeartRateDriftPercent *float64    `json:"heart_rate_drift_percent,omitempty"`
	Efficiency            *Efficiency `json:"efficiency,omitempty"`
}

// Efficiency is the aerobic efficiency of the reference ride: how much speed or
// power each heartbeat buys. A higher value on the same Kind is a better
// condition; values of different kinds are not comparable.
type Efficiency struct {
	Value float64 `json:"value"`
	Kind  string  `json:"kind"`
}

type AssessmentStore interface {
	CurrentAssessmentByUserID(context.Context, string) (*Assessment, error)
	AssessmentHistoryByUserID(context.Context, string, int) ([]Assessment, error)
	SaveSubmaxAssessment(context.Context, string, Assessment) (Assessment, error)
}

type AssessmentService struct{ store AssessmentStore }

func NewAssessmentService(store AssessmentStore) *AssessmentService {
	return &AssessmentService{store: store}
}

func (s *AssessmentService) Current(ctx context.Context, userID string) (*Assessment, error) {
	current, err := s.store.CurrentAssessmentByUserID(ctx, userID)
	if current != nil {
		current.Derive()
	}
	return current, err
}

// History lists the latest assessments, newest first, with the derived numbers.
func (s *AssessmentService) History(ctx context.Context, userID string) ([]Assessment, error) {
	history, err := s.store.AssessmentHistoryByUserID(ctx, userID, historyLimit)
	if err != nil {
		return nil, err
	}
	for index := range history {
		history[index].Derive()
	}
	return history, nil
}

func (s *AssessmentService) SaveSubmax(ctx context.Context, userID string, input Assessment) (Assessment, error) {
	if input.DurationMinutes < 15 || input.DurationMinutes > 30 || input.ActualRPE < 1 || input.ActualRPE > 10 || len(input.Notes) > 1000 {
		return Assessment{}, ErrInvalidAssessment
	}
	if !input.validNumbers() {
		return Assessment{}, ErrInvalidAssessment
	}
	input.AssessmentType = "submax_reference"
	input.TargetRPE = 5
	// The rule is the same as before the numbers existed: no pain, at least 18
	// minutes and an effort up to RPE 6 (Z3). The numbers do not take part.
	input.EligibleForProgression = !input.PainReported && input.DurationMinutes >= 18 && input.ActualRPE <= 6
	saved, err := s.store.SaveSubmaxAssessment(ctx, userID, input)
	if err == nil {
		saved.Derive()
	}
	return saved, err
}

func validHeartRate(value *int) bool { return value == nil || (*value >= 30 && *value <= 250) }

func (a Assessment) validNumbers() bool {
	if !validHeartRate(a.AverageHeartRate) || !validHeartRate(a.HeartRateFirstHalf) || !validHeartRate(a.HeartRateSecondHalf) {
		return false
	}
	if (a.HeartRateFirstHalf == nil) != (a.HeartRateSecondHalf == nil) {
		return false
	}
	if a.AveragePowerW != nil && (*a.AveragePowerW < 0 || *a.AveragePowerW > 2000) {
		return false
	}
	if a.DistanceKM != nil {
		distance := *a.DistanceKM
		if math.IsNaN(distance) || distance < 0 || distance > 500 {
			return false
		}
		if distance > maxPlausibleSpeedKMH*float64(a.DurationMinutes)/60 {
			return false
		}
	}
	return true
}

// meanHeartRate is the average the efficiency uses: the reported average, or the
// mean of the two halves when only those were given.
func (a Assessment) meanHeartRate() float64 {
	if a.AverageHeartRate != nil {
		return float64(*a.AverageHeartRate)
	}
	if a.HeartRateFirstHalf != nil && a.HeartRateSecondHalf != nil {
		return float64(*a.HeartRateFirstHalf+*a.HeartRateSecondHalf) / 2
	}
	return 0
}

func round(value float64, places int) float64 {
	scale := math.Pow(10, float64(places))
	return math.Round(value*scale) / scale
}

// Derive fills the heart rate drift and the aerobic efficiency from the stored
// numbers. Both stay nil when the athlete did not give what they need.
func (a *Assessment) Derive() {
	a.HeartRateDriftPercent, a.Efficiency = nil, nil
	if a.HeartRateFirstHalf != nil && a.HeartRateSecondHalf != nil && *a.HeartRateFirstHalf > 0 {
		drift := round(float64(*a.HeartRateSecondHalf-*a.HeartRateFirstHalf)/float64(*a.HeartRateFirstHalf)*100, 1)
		a.HeartRateDriftPercent = &drift
	}
	heartRate := a.meanHeartRate()
	if heartRate <= 0 {
		return
	}
	if a.AveragePowerW != nil && *a.AveragePowerW > 0 {
		a.Efficiency = &Efficiency{Value: round(float64(*a.AveragePowerW)/heartRate, 2), Kind: EfficiencyPowerPerBPM}
		return
	}
	if a.DistanceKM != nil && *a.DistanceKM > 0 && a.DurationMinutes > 0 {
		speed := *a.DistanceKM / (float64(a.DurationMinutes) / 60)
		a.Efficiency = &Efficiency{Value: round(speed/heartRate*100, 2), Kind: EfficiencySpeedPer100BPM}
	}
}
