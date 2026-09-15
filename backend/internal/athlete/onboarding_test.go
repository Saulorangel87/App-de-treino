package athlete

import (
	"context"
	"testing"
	"time"
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

func TestSaveLimitationsAcceptsOptionalSafetyContext(t *testing.T) {
	intensity := 6
	startedOn := "2026-09-01"
	result, err := NewOnboardingService(onboardingStore{}).SaveLimitations(context.Background(), "user-1", []Limitation{{
		Kind: "pain", Description: "Joelho ao subir", Location: "joelho direito", Intensity: &intensity,
		AggravatingMovement: "Subir em pé", StartedOn: &startedOn,
		SymptomsDuringAfter: []string{"dizziness", "extreme_fatigue"}, MedicalRestriction: true,
		RecentSurgery: true, ExerciseProhibited: true, ConditionAffectingExercise: true,
	}})
	if err != nil || result[0].Location != "joelho direito" || result[0].Intensity == nil || *result[0].Intensity != intensity || result[0].StartedOn == nil || *result[0].StartedOn != startedOn || len(result[0].SymptomsDuringAfter) != 2 || !result[0].MedicalRestriction || !result[0].RecentSurgery || !result[0].ExerciseProhibited || !result[0].ConditionAffectingExercise {
		t.Fatalf("expected optional safety context to be preserved, got %#v, %v", result, err)
	}
}

func TestSaveLimitationsRejectsInvalidOptionalSafetyContext(t *testing.T) {
	intensity := 11
	if _, err := NewOnboardingService(onboardingStore{}).SaveLimitations(context.Background(), "user-1", []Limitation{{Kind: "pain", Description: "Dor", Intensity: &intensity}}); err != ErrInvalidOnboarding {
		t.Fatalf("expected out-of-range intensity to be rejected, got %v", err)
	}
	future := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	if _, err := NewOnboardingService(onboardingStore{}).SaveLimitations(context.Background(), "user-1", []Limitation{{Kind: "pain", Description: "Dor", StartedOn: &future}}); err != ErrInvalidOnboarding {
		t.Fatalf("expected future start date to be rejected, got %v", err)
	}
	if _, err := NewOnboardingService(onboardingStore{}).SaveLimitations(context.Background(), "user-1", []Limitation{{Kind: "pain", Description: "Dor", SymptomsDuringAfter: []string{"dizziness", "dizziness"}}}); err != ErrInvalidOnboarding {
		t.Fatalf("expected duplicate safety symptom to be rejected, got %v", err)
	}
	if _, err := NewOnboardingService(onboardingStore{}).SaveLimitations(context.Background(), "user-1", []Limitation{{Kind: "pain", Description: "Dor", SymptomsDuringAfter: []string{"chest_pain"}}}); err != ErrInvalidOnboarding {
		t.Fatalf("expected unknown safety symptom to be rejected, got %v", err)
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
	result, err := NewOnboardingService(onboardingStore{}).SaveCyclingContext(context.Background(), "user-1", CyclingContext{FTP: &ftp})
	if err != nil || result.FTP != nil {
		t.Fatalf("expected FTP without power meter to be cleared, got %#v, %v", result, err)
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
	averagePower := 185
	distance := 100
	date := "2026-11-15"
	ftpTestDate := "2026-09-10"
	input := CyclingContext{WeeklyHours: 6.5, PracticeDurationMonths: 24, AverageRideMinutes: 75, LongestRideMinutes: 180, WeeklyRides: 4, RecentWeeklyDistanceKM: 160, RecentTrainingWeeks: 12, RecentBestDistanceKM: 95, PreferredSessionTypes: []string{"cadence", "hills"}, Discipline: "road", BikeType: "road", Terrain: "hilly", UsesHeartRate: true, UsesPower: true, UsesGPS: true, UsesSportsWatch: true, UsesSmartTrainer: true, FTP: &ftp, FTPTestDate: &ftpTestDate, FTPProtocol: "20_minute", AveragePowerWatts: &averagePower, EventGoal: true, EventDistanceKM: &distance, EventDate: &date}
	service := NewOnboardingService(onboardingStore{})
	service.now = func() time.Time { return time.Date(2026, time.September, 11, 12, 0, 0, 0, time.Local) }
	result, err := service.SaveCyclingContext(context.Background(), "user-1", input)
	if err != nil || result.Discipline != "road" || result.FTP == nil || *result.FTP != ftp || result.FTPTestDate == nil || *result.FTPTestDate != ftpTestDate || result.AveragePowerWatts == nil || *result.AveragePowerWatts != averagePower || !result.UsesGPS || !result.UsesSportsWatch || !result.UsesSmartTrainer || result.EventDate == nil || *result.EventDate != date {
		t.Fatalf("expected valid cycling context, got %#v, %v", result, err)
	}
}

func TestSaveCyclingContextRejectsFutureFTPTestDate(t *testing.T) {
	ftpTestDate := "2026-09-12"
	service := NewOnboardingService(onboardingStore{})
	service.now = func() time.Time { return time.Date(2026, time.September, 11, 12, 0, 0, 0, time.Local) }
	if _, err := service.SaveCyclingContext(context.Background(), "user-1", CyclingContext{UsesPower: true, FTPTestDate: &ftpTestDate}); err != ErrInvalidOnboarding {
		t.Fatalf("expected future FTP test date to be rejected, got %v", err)
	}
}

func TestSaveCyclingContextClearsPowerDetailsWithoutMeter(t *testing.T) {
	ftp := 220
	ftpTestDate := "2026-09-10"
	averagePower := 185
	result, err := NewOnboardingService(onboardingStore{}).SaveCyclingContext(context.Background(), "user-1", CyclingContext{FTP: &ftp, FTPTestDate: &ftpTestDate, FTPProtocol: "20_minute", AveragePowerWatts: &averagePower})
	if err != nil || result.FTP != nil || result.FTPTestDate != nil || result.AveragePowerWatts != nil || result.FTPProtocol != "" {
		t.Fatalf("expected power details without meter to be cleared, got %#v, %v", result, err)
	}
}

func TestSaveCyclingContextRejectsPastEventDate(t *testing.T) {
	ftp := 220
	distance := 100
	past := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	input := CyclingContext{UsesPower: true, FTP: &ftp, EventGoal: true, EventDistanceKM: &distance, EventDate: &past}
	if _, err := NewOnboardingService(onboardingStore{}).SaveCyclingContext(context.Background(), "user-1", input); err != ErrInvalidOnboarding {
		t.Fatalf("expected past event date to be rejected, got %v", err)
	}
}

func TestSaveCyclingContextAcceptsTodayEventDate(t *testing.T) {
	ftp := 220
	distance := 100
	today := "2026-09-11"
	input := CyclingContext{UsesPower: true, FTP: &ftp, EventGoal: true, EventDistanceKM: &distance, EventDate: &today}
	service := NewOnboardingService(onboardingStore{})
	service.now = func() time.Time { return time.Date(2026, time.September, 11, 12, 0, 0, 0, time.Local) }
	if _, err := service.SaveCyclingContext(context.Background(), "user-1", input); err != nil {
		t.Fatalf("expected today's event date to be accepted, got %v", err)
	}
}

func TestSaveCyclingContextAcceptsKnownDisciplines(t *testing.T) {
	for _, discipline := range []string{"", "general", "road", "mtb_xco", "mtb_xcm", "gravel", "indoor"} {
		result, err := NewOnboardingService(onboardingStore{}).SaveCyclingContext(context.Background(), "user-1", CyclingContext{Discipline: " " + discipline + " "})
		if err != nil || result.Discipline != discipline {
			t.Fatalf("expected discipline %q to be accepted and normalized, got %#v, %v", discipline, result, err)
		}
	}
}

func TestSaveCyclingContextRejectsExcludedDisciplines(t *testing.T) {
	for _, discipline := range []string{"dh_enduro", "track_sprint"} {
		if _, err := NewOnboardingService(onboardingStore{}).SaveCyclingContext(context.Background(), "user-1", CyclingContext{Discipline: discipline}); err != ErrInvalidOnboarding {
			t.Fatalf("expected excluded discipline %q to be rejected, got %v", discipline, err)
		}
	}
}

func TestSaveCyclingContextAcceptsAllSessionPreferences(t *testing.T) {
	input := CyclingContext{PreferredSessionTypes: []string{"base", "cadence", "hills", "intervals", "threshold", "sweet_spot", "vo2max", "short_intervals", "recovery"}}
	if _, err := NewOnboardingService(onboardingStore{}).SaveCyclingContext(context.Background(), "user-1", input); err != nil {
		t.Fatalf("expected all available session preferences to be accepted, got %v", err)
	}
}

func TestSaveCyclingContextAcceptsShortIntervalsSessionPreference(t *testing.T) {
	input := CyclingContext{PreferredSessionTypes: []string{"short_intervals"}}
	if _, err := NewOnboardingService(onboardingStore{}).SaveCyclingContext(context.Background(), "user-1", input); err != nil {
		t.Fatalf("expected short intervals session preference to be accepted, got %v", err)
	}
}

func TestSaveCyclingContextAcceptsVO2MaxSessionPreference(t *testing.T) {
	input := CyclingContext{PreferredSessionTypes: []string{"vo2max"}}
	if _, err := NewOnboardingService(onboardingStore{}).SaveCyclingContext(context.Background(), "user-1", input); err != nil {
		t.Fatalf("expected VO₂max session preference to be accepted, got %v", err)
	}
}

func TestSaveCyclingContextAcceptsThresholdSessionPreference(t *testing.T) {
	input := CyclingContext{PreferredSessionTypes: []string{"threshold"}}
	if _, err := NewOnboardingService(onboardingStore{}).SaveCyclingContext(context.Background(), "user-1", input); err != nil {
		t.Fatalf("expected threshold session preference to be accepted, got %v", err)
	}
}

func TestSaveCyclingContextRejectsInvalidHistory(t *testing.T) {
	service := NewOnboardingService(onboardingStore{})
	if _, err := service.SaveCyclingContext(context.Background(), "user-1", CyclingContext{WeeklyRides: 22}); err != ErrInvalidOnboarding {
		t.Fatalf("expected weekly ride history to be bounded, got %v", err)
	}
	if _, err := service.SaveCyclingContext(context.Background(), "user-1", CyclingContext{RecentWeeklyDistanceKM: 2001}); err != ErrInvalidOnboarding {
		t.Fatalf("expected weekly distance history to be bounded, got %v", err)
	}
	if _, err := service.SaveCyclingContext(context.Background(), "user-1", CyclingContext{RecentTrainingWeeks: 53}); err != ErrInvalidOnboarding {
		t.Fatalf("expected training history duration to be bounded, got %v", err)
	}
	if _, err := service.SaveCyclingContext(context.Background(), "user-1", CyclingContext{RecentBestDistanceKM: 2001}); err != ErrInvalidOnboarding {
		t.Fatalf("expected best distance history to be bounded, got %v", err)
	}
	if _, err := service.SaveCyclingContext(context.Background(), "user-1", CyclingContext{PreferredSessionTypes: []string{"unknown"}}); err != ErrInvalidOnboarding {
		t.Fatalf("expected unknown session preference to be rejected, got %v", err)
	}
	if _, err := service.SaveCyclingContext(context.Background(), "user-1", CyclingContext{Discipline: "mtb"}); err != ErrInvalidOnboarding {
		t.Fatalf("expected unknown discipline to be rejected, got %v", err)
	}
}

func TestSaveCyclingContextDoesNotInferDisciplineFromBikeType(t *testing.T) {
	result, err := NewOnboardingService(onboardingStore{}).SaveCyclingContext(context.Background(), "user-1", CyclingContext{BikeType: "mtb"})
	if err != nil {
		t.Fatalf("expected bike type without discipline to remain valid, got %v", err)
	}
	if result.Discipline != "" {
		t.Fatalf("expected discipline to remain unset instead of being inferred, got %q", result.Discipline)
	}
}

func TestSaveCyclingContextNormalizesMissingPreferences(t *testing.T) {
	result, err := NewOnboardingService(onboardingStore{}).SaveCyclingContext(context.Background(), "user-1", CyclingContext{})
	if err != nil {
		t.Fatalf("expected empty optional context to be accepted, got %v", err)
	}
	if result.PreferredSessionTypes == nil {
		t.Fatal("expected missing preferences to be normalized to an empty list")
	}
	if result.TrainingStatus != "not_informed" {
		t.Fatalf("expected missing training status to be normalized, got %q", result.TrainingStatus)
	}
}

func TestSaveCyclingContextAcceptsTrainingStatuses(t *testing.T) {
	for _, status := range []string{"not_informed", "regular", "returning_after_break"} {
		result, err := NewOnboardingService(onboardingStore{}).SaveCyclingContext(context.Background(), "user-1", CyclingContext{TrainingStatus: " " + status + " "})
		if err != nil || result.TrainingStatus != status {
			t.Fatalf("expected training status %q to be accepted and normalized, got %#v, %v", status, result, err)
		}
	}
}

func TestSaveCyclingContextRejectsUnknownTrainingStatus(t *testing.T) {
	if _, err := NewOnboardingService(onboardingStore{}).SaveCyclingContext(context.Background(), "user-1", CyclingContext{TrainingStatus: "paused"}); err != ErrInvalidOnboarding {
		t.Fatalf("expected unknown training status to be rejected, got %v", err)
	}
}

func errorsIs(err, target error) bool { return err == target }
