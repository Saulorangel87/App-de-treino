package athlete

import (
	"testing"
	"time"
)

func TestRecoveryReadinessIsConservative(t *testing.T) {
	tests := []struct {
		name string
		in   RecoveryCheckin
		want string
	}{
		{"ready", RecoveryCheckin{SleepMinutes: 480, SleepQuality: 4, StressLevel: 2, FatigueLevel: 2}, "ready"},
		{"single warning", RecoveryCheckin{SleepMinutes: 480, SleepQuality: 2, StressLevel: 2, FatigueLevel: 2}, "caution"},
		{"short sleep", RecoveryCheckin{SleepMinutes: 330, SleepQuality: 4, StressLevel: 2, FatigueLevel: 2}, "caution"},
		{"combined warning", RecoveryCheckin{SleepMinutes: 480, SleepQuality: 2, StressLevel: 4, FatigueLevel: 3}, "recovery_needed"},
		{"maximum fatigue", RecoveryCheckin{SleepMinutes: 480, SleepQuality: 4, StressLevel: 2, FatigueLevel: 5}, "recovery_needed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := RecoveryReadiness(test.in); got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestRecoveryDateAllowsClientTimezoneBoundary(t *testing.T) {
	now := time.Date(2026, 9, 1, 1, 0, 0, 0, time.UTC)
	if !validRecoveryDateFor("2026-08-31", now) || !validRecoveryDateFor("2026-09-02", now) || validRecoveryDateFor("2026-08-30", now) {
		t.Fatal("unexpected local recovery date boundary")
	}
}

func TestRecoveryDateValidation(t *testing.T) {
	if !validRecoveryDate("2026-08-31") || validRecoveryDate("31/08/2026") || validRecoveryDate("2026-02-30") {
		t.Fatal("unexpected recovery date validation")
	}
}
