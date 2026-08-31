package athlete

import (
	"context"
	"testing"
)

type onboardingStore struct{}

func (onboardingStore) OnboardingByUserID(context.Context, string) (Onboarding, error) {
	return Onboarding{}, nil
}
func (onboardingStore) ReplaceLimitations(_ context.Context, _ string, value []Limitation) ([]Limitation, error) {
	return value, nil
}
func (onboardingStore) ReplaceGoals(_ context.Context, _ string, value []Goal) ([]Goal, error) {
	return value, nil
}
func (onboardingStore) ReplaceAvailability(_ context.Context, _ string, value []Availability) ([]Availability, error) {
	return value, nil
}
func (onboardingStore) SaveCyclingContext(_ context.Context, _ string, value CyclingContext) (CyclingContext, error) {
	return value, nil
}

func TestSaveLimitationsAllowsAthleteWithoutLimitations(t *testing.T) {
	result, err := NewOnboardingService(onboardingStore{}).SaveLimitations(context.Background(), "user-1", []Limitation{})
	if err != nil || len(result) != 0 {
		t.Fatalf("expected empty limitations, got %#v, %v", result, err)
	}
}

func TestSaveGoalsRequiresPrimaryGoal(t *testing.T) {
	_, err := NewOnboardingService(onboardingStore{}).SaveGoals(context.Background(), "user-1", []Goal{{GoalType: "health", Priority: 2}})
	if !errorsIs(err, ErrInvalidOnboarding) {
		t.Fatalf("expected ErrInvalidOnboarding, got %v", err)
	}
}

func TestSaveAvailabilityRequiresAtLeastOneTrainingDay(t *testing.T) {
	items := make([]Availability, 7)
	for index := range items {
		items[index].Weekday = index
	}
	_, err := NewOnboardingService(onboardingStore{}).SaveAvailability(context.Background(), "user-1", items)
	if !errorsIs(err, ErrInvalidOnboarding) {
		t.Fatalf("expected ErrInvalidOnboarding, got %v", err)
	}
}

func TestSaveAvailabilityAllowsUpToEightHoursPerDay(t *testing.T) {
	items := make([]Availability, 7)
	for index := range items {
		items[index].Weekday = index
	}
	items[6].AvailableMinutes = 480
	if _, err := NewOnboardingService(onboardingStore{}).SaveAvailability(context.Background(), "user-1", items); err != nil {
		t.Fatalf("expected eight hours to be accepted, got %v", err)
	}
	items[6].AvailableMinutes = 481
	if _, err := NewOnboardingService(onboardingStore{}).SaveAvailability(context.Background(), "user-1", items); !errorsIs(err, ErrInvalidOnboarding) {
		t.Fatalf("expected more than eight hours to be rejected, got %v", err)
	}
}

func TestSaveCyclingContextRequiresPowerMeterForFTP(t *testing.T) {
	ftp := 220
	_, err := NewOnboardingService(onboardingStore{}).SaveCyclingContext(context.Background(), "user-1", CyclingContext{FTP: &ftp})
	if !errorsIs(err, ErrInvalidOnboarding) {
		t.Fatalf("expected FTP without power meter to be rejected, got %v", err)
	}
}

func TestSaveCyclingContextRequiresCompleteEventGoal(t *testing.T) {
	_, err := NewOnboardingService(onboardingStore{}).SaveCyclingContext(context.Background(), "user-1", CyclingContext{EventGoal: true})
	if !errorsIs(err, ErrInvalidOnboarding) {
		t.Fatalf("expected incomplete event goal to be rejected, got %v", err)
	}
}

func TestSaveCyclingContextAcceptsOptionalContext(t *testing.T) {
	ftp := 220
	distance := 100
	date := "2026-11-15"
	input := CyclingContext{WeeklyHours: 6.5, LongestRideMinutes: 180, BikeType: "road", Terrain: "hilly", UsesHeartRate: true, UsesPower: true, FTP: &ftp, EventGoal: true, EventDistanceKM: &distance, EventDate: &date}
	result, err := NewOnboardingService(onboardingStore{}).SaveCyclingContext(context.Background(), "user-1", input)
	if err != nil || result.FTP == nil || *result.FTP != ftp || result.EventDate == nil || *result.EventDate != date {
		t.Fatalf("expected valid cycling context, got %#v, %v", result, err)
	}
}

func errorsIs(err, target error) bool { return err == target }
