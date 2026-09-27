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
	hashes     map[string]bool
	saved      []Activity
	saveErr    error
	candidates []WorkoutCandidate
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

func (f *fakeStore) WorkoutCandidatesOnDate(_ context.Context, _ string, _ time.Time) ([]WorkoutCandidate, error) {
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
