package activityimport

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
)

type fakeStore struct {
	hashes       map[string]bool
	saved        []Activity
	saveErr      error
	candidates   []WorkoutCandidate
	queriedDates []string
	// ownedWorkouts são os treinos que pertencem ao atleta no teste de vínculo.
	ownedWorkouts map[string]bool
	linkCalls     int
}

func newFakeStore() *fakeStore { return &fakeStore{hashes: map[string]bool{}} }

func (f *fakeStore) ActivityExists(_ context.Context, _, fileHash string) (bool, error) {
	return f.hashes[fileHash], nil
}

func (f *fakeStore) SaveActivity(_ context.Context, userID, source, fileHash string, parsed Parsed) (Activity, error) {
	if f.saveErr != nil {
		return Activity{}, f.saveErr
	}
	f.hashes[fileHash] = true
	activity := Activity{ID: "activity-1", UserID: userID, Parsed: parsed, Source: source, ImportedAt: time.Now()}
	f.saved = append(f.saved, activity)
	return activity, nil
}

func (f *fakeStore) ListActivities(_ context.Context, _ string) ([]Activity, error) {
	return f.saved, nil
}

func (f *fakeStore) DeleteActivity(_ context.Context, _, _ string) error { return nil }

func (f *fakeStore) GetActivity(_ context.Context, _, activityID string) (Activity, error) {
	for _, activity := range f.saved {
		if activity.ID == activityID {
			return activity, nil
		}
	}
	return Activity{}, ErrNotFound
}

func (f *fakeStore) LinkActivity(_ context.Context, _, activityID string, workoutID *string) error {
	f.linkCalls++
	if workoutID != nil && !f.ownedWorkouts[*workoutID] {
		return ErrWorkoutNotFound
	}
	for i := range f.saved {
		if f.saved[i].ID == activityID {
			f.saved[i].WorkoutID = workoutID
		}
	}
	return nil
}

func (f *fakeStore) WorkoutCandidatesOnDate(_ context.Context, _ string, date time.Time) ([]WorkoutCandidate, error) {
	f.queriedDates = append(f.queriedDates, date.Format("2006-01-02"))
	return f.candidates, nil
}

func sampleGPXBytes() []byte { return []byte(sampleGPX) }

func TestService_Import_StoresAndSuggestsCandidate(t *testing.T) {
	store := newFakeStore()
	store.candidates = []WorkoutCandidate{{ID: "w1", ScheduledOn: "2026-09-20", Name: "Sweet spot", Status: "planned"}}
	service := NewService(store)

	activity, candidates, err := service.Import(context.Background(), "user-1", "ride.gpx", bytes.NewReader(sampleGPXBytes()))
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if activity.Source != SourceGPX {
		t.Errorf("Source = %q, want %q", activity.Source, SourceGPX)
	}
	if len(candidates) != 1 || candidates[0].ID != "w1" {
		t.Errorf("candidates = %+v, want one candidate w1", candidates)
	}
}

func TestService_Import_RejectsDuplicateFile(t *testing.T) {
	store := newFakeStore()
	service := NewService(store)
	ctx := context.Background()

	if _, _, err := service.Import(ctx, "user-1", "ride.gpx", bytes.NewReader(sampleGPXBytes())); err != nil {
		t.Fatalf("first Import: %v", err)
	}
	_, _, err := service.Import(ctx, "user-1", "ride.gpx", bytes.NewReader(sampleGPXBytes()))
	if err != ErrDuplicate {
		t.Fatalf("second Import err = %v, want ErrDuplicate", err)
	}
}

func TestService_Import_RejectsUnrecognizedExtension(t *testing.T) {
	service := NewService(newFakeStore())
	_, _, err := service.Import(context.Background(), "user-1", "ride.tcx", strings.NewReader("whatever"))
	if err != ErrUnrecognizedFormat {
		t.Fatalf("err = %v, want ErrUnrecognizedFormat", err)
	}
}

func TestService_Import_RejectsFileOverLimit(t *testing.T) {
	service := NewService(newFakeStore())
	oversized := bytes.Repeat([]byte("a"), MaxFileSizeBytes+1)
	_, _, err := service.Import(context.Background(), "user-1", "ride.gpx", bytes.NewReader(oversized))
	if err != ErrFileTooLarge {
		t.Fatalf("err = %v, want ErrFileTooLarge", err)
	}
}

// TestService_Import_NeverTouchesPrescription documents the invariant the
// whole feature depends on: importing an activity only produces a Parsed
// summary and, at most, a suggested WorkoutCandidate. It never calls into
// the planning package and never writes to a workout or training plan —
// only the athlete, via the existing completion/correction endpoints, can
// turn an import into recorded execution data.
func TestService_Import_NeverTouchesPrescription(t *testing.T) {
	store := newFakeStore()
	service := NewService(store)
	if _, _, err := service.Import(context.Background(), "user-1", "ride.gpx", bytes.NewReader(sampleGPXBytes())); err != nil {
		t.Fatalf("Import: %v", err)
	}
	if len(store.saved) != 1 {
		t.Fatalf("expected exactly one saved activity, got %d", len(store.saved))
	}
	if store.saved[0].WorkoutID != nil {
		t.Fatalf("SaveActivity must not receive a workout link from Import; got %v", store.saved[0].WorkoutID)
	}
}

func buildFITBytes(t *testing.T) []byte {
	t.Helper()
	session := mesgdef.NewSession(nil)
	session.Sport = typedef.SportCycling
	session.StartTime = time.Now().UTC()
	session.TotalTimerTime = 1800 * 1000
	session.TotalElapsedTime = 1800 * 1000
	return buildFIT(t, session.ToMesg(nil))
}

func TestService_Import_AcceptsFIT(t *testing.T) {
	service := NewService(newFakeStore())
	activity, _, err := service.Import(context.Background(), "user-1", "RIDE.FIT", bytes.NewReader(buildFITBytes(t)))
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if activity.Source != SourceFIT {
		t.Errorf("Source = %q, want %q (extension match must be case-insensitive)", activity.Source, SourceFIT)
	}
}

// TestService_Import_WidensSearchWithoutKnownTimezone documents the fix for
// a ride close to local midnight: without a way to tell the athlete's local
// calendar day (always the case for .gpx), the exact UTC date can miss the
// planned workout entirely. The service must check the day before and after
// too, instead of matching only the UTC-derived date.
func TestService_Import_WidensSearchWithoutKnownTimezone(t *testing.T) {
	store := newFakeStore()
	service := NewService(store)
	if _, _, err := service.Import(context.Background(), "user-1", "ride.gpx", bytes.NewReader(sampleGPXBytes())); err != nil {
		t.Fatalf("Import: %v", err)
	}
	if len(store.queriedDates) != 3 {
		t.Fatalf("queriedDates = %v, want 3 dates (day-1, day, day+1)", store.queriedDates)
	}
}

// TestService_Import_ExactSearchWithKnownTimezone documents the other half
// of the same fix: a .fit file that does carry the device's local timezone
// (the `activity` message's LocalTimestamp) should not need the widened,
// approximate search — the exact local calendar date is already correct.
func TestService_Import_ExactSearchWithKnownTimezone(t *testing.T) {
	utcStart := time.Date(2026, 9, 20, 23, 30, 0, 0, time.UTC) // 20:30 in UTC-3, still the 20th locally
	session := mesgdef.NewSession(nil)
	session.Sport = typedef.SportCycling
	session.StartTime = utcStart
	session.TotalTimerTime = 1800 * 1000
	session.TotalElapsedTime = 1800 * 1000
	activityMsg := mesgdef.NewActivity(nil)
	activityMsg.Timestamp = utcStart
	activityMsg.LocalTimestamp = utcStart.Add(-3 * time.Hour)
	data := buildFIT(t, session.ToMesg(nil), activityMsg.ToMesg(nil))

	store := newFakeStore()
	service := NewService(store)
	activity, _, err := service.Import(context.Background(), "user-1", "ride.fit", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if len(store.queriedDates) != 1 {
		t.Fatalf("queriedDates = %v, want exactly 1 (the known local date)", store.queriedDates)
	}
	if want := "2026-09-20"; store.queriedDates[0] != want {
		t.Errorf("queried date = %q, want %q (local date, not the UTC date 2026-09-21)", store.queriedDates[0], want)
	}
	if activity.StartedAt.Hour() != 20 {
		t.Errorf("StartedAt hour = %d, want 20 (23:30 UTC shifted by -3h)", activity.StartedAt.Hour())
	}
}

const validWorkoutID = "11111111-2222-3333-4444-555555555555"

func storeWithImportedActivity(t *testing.T) (*fakeStore, *Service) {
	t.Helper()
	store := newFakeStore()
	store.ownedWorkouts = map[string]bool{validWorkoutID: true}
	service := NewService(store)
	if _, _, err := service.Import(context.Background(), "user-1", "ride.gpx", bytes.NewReader(sampleGPXBytes())); err != nil {
		t.Fatalf("Import: %v", err)
	}
	return store, service
}

func TestService_Link_LinksAndUnlinks(t *testing.T) {
	_, service := storeWithImportedActivity(t)
	workoutID := validWorkoutID

	linked, err := service.Link(context.Background(), "user-1", "activity-1", &workoutID)
	if err != nil {
		t.Fatalf("Link: %v", err)
	}
	if linked.WorkoutID == nil || *linked.WorkoutID != validWorkoutID {
		t.Fatalf("WorkoutID = %v, want %s", linked.WorkoutID, validWorkoutID)
	}

	unlinked, err := service.Link(context.Background(), "user-1", "activity-1", nil)
	if err != nil {
		t.Fatalf("Link(nil): %v", err)
	}
	if unlinked.WorkoutID != nil {
		t.Fatalf("WorkoutID = %v, want nil after unlinking", unlinked.WorkoutID)
	}
}

func TestService_Link_RejectsUnknownActivity(t *testing.T) {
	store, service := storeWithImportedActivity(t)
	workoutID := validWorkoutID
	if _, err := service.Link(context.Background(), "user-1", "missing", &workoutID); err != ErrNotFound {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	if store.linkCalls != 0 {
		t.Fatalf("LinkActivity called %d times for an unknown activity", store.linkCalls)
	}
}

func TestService_Link_RejectsWorkoutOfAnotherAthlete(t *testing.T) {
	_, service := storeWithImportedActivity(t)
	other := "99999999-2222-3333-4444-555555555555"
	if _, err := service.Link(context.Background(), "user-1", "activity-1", &other); err != ErrWorkoutNotFound {
		t.Fatalf("err = %v, want ErrWorkoutNotFound", err)
	}
}

func TestService_Link_RejectsMalformedWorkoutID(t *testing.T) {
	store, service := storeWithImportedActivity(t)
	for _, bad := range []string{"", "abc", "11111111-2222-3333-4444-55555555555", "' OR 1=1 --"} {
		id := bad
		if _, err := service.Link(context.Background(), "user-1", "activity-1", &id); err != ErrInvalidWorkoutID {
			t.Errorf("Link(%q) err = %v, want ErrInvalidWorkoutID", bad, err)
		}
	}
	if store.linkCalls != 0 {
		t.Fatalf("LinkActivity must not run for malformed ids; calls = %d", store.linkCalls)
	}
}

func TestService_Candidates_SearchesAdjacentDays(t *testing.T) {
	store, service := storeWithImportedActivity(t)
	store.queriedDates = nil
	store.candidates = []WorkoutCandidate{{ID: validWorkoutID, ScheduledOn: "2026-09-20", Name: "Giro de base", Status: "planned"}}

	candidates, err := service.Candidates(context.Background(), "user-1", "activity-1")
	if err != nil {
		t.Fatalf("Candidates: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("candidates = %+v, want one (deduplicated across days)", candidates)
	}
	want := map[string]bool{"2026-09-20": true, "2026-09-19": true, "2026-09-21": true}
	if len(store.queriedDates) != 3 {
		t.Fatalf("queried %v, want the day and its neighbors", store.queriedDates)
	}
	for _, day := range store.queriedDates {
		if !want[day] {
			t.Errorf("unexpected queried day %s", day)
		}
	}
}

func TestService_Candidates_UnknownActivity(t *testing.T) {
	_, service := storeWithImportedActivity(t)
	if _, err := service.Candidates(context.Background(), "user-1", "missing"); err != ErrNotFound {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
