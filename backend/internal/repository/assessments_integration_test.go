package repository

import (
	"testing"

	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
)

// Runs only when CADENCIA_TEST_DATABASE_URL points at a migrated database (000037).

func intPtr(value int) *int { return &value }

func TestAssessmentsKeepTheNumbersAndListNewestFirst(t *testing.T) {
	f := newUndoFixture(t)
	distance := 9.5
	service := athlete.NewAssessmentService(f.store)

	first, err := service.SaveSubmax(f.ctx, f.userID, athlete.Assessment{DurationMinutes: 20, ActualRPE: 4.5})
	if err != nil {
		t.Fatalf("save without numbers: %v", err)
	}
	if first.Efficiency != nil || first.AverageHeartRate != nil {
		t.Fatalf("an assessment without numbers must come back without them: %+v", first)
	}
	f.exec(`UPDATE cycling_assessments SET completed_at = now() - interval '30 days' WHERE id = $1`, first.ID)

	second, err := service.SaveSubmax(f.ctx, f.userID, athlete.Assessment{
		DurationMinutes: 20, ActualRPE: 4.5, AverageHeartRate: intPtr(140), AveragePowerW: intPtr(168), DistanceKM: &distance,
		HeartRateFirstHalf: intPtr(135), HeartRateSecondHalf: intPtr(144),
	})
	if err != nil {
		t.Fatalf("save with numbers: %v", err)
	}
	if !second.EligibleForProgression || second.Efficiency == nil || second.Efficiency.Value != 1.2 {
		t.Fatalf("saved assessment lost or miscalculated the numbers: %+v", second)
	}

	history, err := service.History(f.ctx, f.userID)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(history) != 2 || history[0].ID != second.ID || history[1].ID != first.ID {
		t.Fatalf("history must list the newest first, got %+v", history)
	}
	got := history[0]
	if got.AverageHeartRate == nil || *got.AverageHeartRate != 140 || got.DistanceKM == nil || *got.DistanceKM != 9.5 ||
		got.HeartRateDriftPercent == nil || *got.HeartRateDriftPercent != 6.7 {
		t.Fatalf("stored numbers did not survive the round trip: %+v", got)
	}

	current, err := service.Current(f.ctx, f.userID)
	if err != nil || current == nil || current.ID != second.ID {
		t.Fatalf("current must be the newest assessment: %+v %v", current, err)
	}
}
