package repository

import "testing"

func TestAdaptedWorkoutIsStartable(t *testing.T) {
	for _, status := range []string{"planned", "adapted"} {
		if !isStartableWorkoutStatus(status) {
			t.Fatalf("status %q must be startable", status)
		}
	}
	for _, status := range []string{"in_progress", "completed", "skipped"} {
		if isStartableWorkoutStatus(status) {
			t.Fatalf("status %q must not be startable", status)
		}
	}
}
