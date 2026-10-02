package athlete

import (
	"context"
	"errors"
	"testing"
)

type fakeAssessmentStore struct {
	saved Assessment
	err   error
}

func (f *fakeAssessmentStore) CurrentAssessmentByUserID(context.Context, string) (*Assessment, error) {
	return nil, nil
}
func (f *fakeAssessmentStore) AssessmentHistoryByUserID(context.Context, string, int) ([]Assessment, error) {
	return nil, nil
}
func (f *fakeAssessmentStore) SaveSubmaxAssessment(_ context.Context, _ string, input Assessment) (Assessment, error) {
	f.saved = input
	return input, f.err
}

func ptr[T any](value T) *T { return &value }

func TestEligibilityRuleIsUnchangedByTheNumbers(t *testing.T) {
	cases := []struct {
		name  string
		input Assessment
		want  bool
	}{
		{"sem dor, 20 min, Z2", Assessment{DurationMinutes: 20, ActualRPE: 4.5}, true},
		{"Z3 ainda vale (RPE 6)", Assessment{DurationMinutes: 20, ActualRPE: 6}, true},
		{"Z4 não vale", Assessment{DurationMinutes: 20, ActualRPE: 7}, false},
		{"menos de 18 minutos", Assessment{DurationMinutes: 15, ActualRPE: 4.5}, false},
		{"com dor", Assessment{DurationMinutes: 20, ActualRPE: 4.5, PainReported: true}, false},
		{"números ótimos não salvam um esforço forte", Assessment{DurationMinutes: 20, ActualRPE: 8.5, AverageHeartRate: ptr(120), AveragePowerW: ptr(300)}, false},
		{"números ruins não derrubam um pedal certo", Assessment{DurationMinutes: 20, ActualRPE: 4.5, AverageHeartRate: ptr(190), HeartRateFirstHalf: ptr(150), HeartRateSecondHalf: ptr(190)}, true},
	}
	for _, c := range cases {
		store := &fakeAssessmentStore{}
		got, err := NewAssessmentService(store).SaveSubmax(context.Background(), "user", c.input)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got.EligibleForProgression != c.want {
			t.Errorf("%s: eligible=%t, want %t", c.name, got.EligibleForProgression, c.want)
		}
	}
}

func TestDeriveComputesDriftAndEfficiency(t *testing.T) {
	power := Assessment{DurationMinutes: 20, AverageHeartRate: ptr(140), AveragePowerW: ptr(168), HeartRateFirstHalf: ptr(135), HeartRateSecondHalf: ptr(144)}
	power.Derive()
	if power.Efficiency == nil || power.Efficiency.Kind != EfficiencyPowerPerBPM || power.Efficiency.Value != 1.2 {
		t.Errorf("power efficiency = %+v, want 1.2 W/bpm", power.Efficiency)
	}
	if power.HeartRateDriftPercent == nil || *power.HeartRateDriftPercent != 6.7 {
		t.Errorf("drift = %v, want 6.7", power.HeartRateDriftPercent)
	}

	// 10 km in 20 min is 30 km/h; at 150 bpm that is 20 km/h per 100 bpm.
	speed := Assessment{DurationMinutes: 20, AverageHeartRate: ptr(150), DistanceKM: ptr(10.0)}
	speed.Derive()
	if speed.Efficiency == nil || speed.Efficiency.Kind != EfficiencySpeedPer100BPM || speed.Efficiency.Value != 20 {
		t.Errorf("speed efficiency = %+v, want 20", speed.Efficiency)
	}
	if speed.HeartRateDriftPercent != nil {
		t.Errorf("drift needs both halves, got %v", *speed.HeartRateDriftPercent)
	}

	// Only the halves: the efficiency uses their mean.
	halves := Assessment{DurationMinutes: 20, AveragePowerW: ptr(150), HeartRateFirstHalf: ptr(130), HeartRateSecondHalf: ptr(140)}
	halves.Derive()
	if halves.Efficiency == nil || halves.Efficiency.Value != 1.11 {
		t.Errorf("efficiency from halves = %+v, want 1.11", halves.Efficiency)
	}

	none := Assessment{DurationMinutes: 20, ActualRPE: 4.5}
	none.Derive()
	if none.Efficiency != nil || none.HeartRateDriftPercent != nil {
		t.Errorf("an assessment without numbers must derive nothing: %+v", none)
	}
}

func TestSaveSubmaxRejectsInconsistentNumbers(t *testing.T) {
	base := func() Assessment { return Assessment{DurationMinutes: 20, ActualRPE: 4.5} }
	bad := map[string]func(*Assessment){
		"uma metade sozinha":     func(a *Assessment) { a.HeartRateFirstHalf = ptr(130) },
		"FC baixa demais":        func(a *Assessment) { a.AverageHeartRate = ptr(20) },
		"FC alta demais":         func(a *Assessment) { a.AverageHeartRate = ptr(300) },
		"potência absurda":       func(a *Assessment) { a.AveragePowerW = ptr(2500) },
		"potência negativa":      func(a *Assessment) { a.AveragePowerW = ptr(-1) },
		"distância negativa":     func(a *Assessment) { a.DistanceKM = ptr(-2.0) },
		"distância que não cabe": func(a *Assessment) { a.DistanceKM = ptr(40.0) },
		"duração de 5 minutos":   func(a *Assessment) { a.DurationMinutes = 5 },
	}
	for name, mutate := range bad {
		input := base()
		mutate(&input)
		store := &fakeAssessmentStore{}
		if _, err := NewAssessmentService(store).SaveSubmax(context.Background(), "user", input); !errors.Is(err, ErrInvalidAssessment) {
			t.Errorf("%s: returned %v, want ErrInvalidAssessment", name, err)
		}
		if store.saved.DurationMinutes != 0 {
			t.Errorf("%s: a refused assessment reached the store", name)
		}
	}
}
