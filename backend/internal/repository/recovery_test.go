package repository

import (
	"strings"
	"testing"

	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
)

func TestRecoveryAdjustmentUsesExplicitTextArguments(t *testing.T) {
	input := athlete.RecoveryCheckin{
		RecordedOn: "2026-09-04",
		Readiness:  "recovery_needed",
	}
	args := recoveryAdjustmentArgs("profile-1", input, 0.8, 4)

	if len(args) != 6 {
		t.Fatalf("got %d query arguments, want 6", len(args))
	}
	if args[1] != input.RecordedOn || args[5] != input.RecordedOn {
		t.Fatalf("date arguments must preserve the recorded date, got %#v", args)
	}
	for _, fragment := range []string{
		") ? ($6::text)",
		"'readiness', $5::text",
		"CASE WHEN $5::text = 'recovery_needed'",
		"jsonb_build_array($6::text)",
	} {
		if !strings.Contains(recoveryAdjustmentQuery, fragment) {
			t.Fatalf("recovery adjustment query is missing explicit text cast %q", fragment)
		}
	}
}
