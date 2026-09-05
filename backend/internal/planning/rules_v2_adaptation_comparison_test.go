package planning

import (
	"testing"
	"time"
)

func TestRulesV2AdaptationShadowMatrixKeepsRulesV1Authoritative(t *testing.T) {
	cases := []struct {
		name              string
		targetRPE         float64
		input             CompletionInput
		periods           []TrainingHistoryPeriod
		rulesV1Kind       string
		shadowStatus      string
		candidateResponse string
	}{
		{
			name:              "pain",
			targetRPE:         6,
			input:             CompletionInput{ActualRPE: 4, Difficulty: "easy", FatigueAfter: 1, PainReported: true},
			periods:           validHistoryPeriods(),
			rulesV1Kind:       "safety",
			shadowStatus:      "protective_signal",
			candidateResponse: "prefer_recovery",
		},
		{
			name:              "high strain",
			targetRPE:         5,
			input:             CompletionInput{ActualRPE: 7, Difficulty: "hard", FatigueAfter: 4},
			periods:           validHistoryPeriods(),
			rulesV1Kind:       "recovery",
			shadowStatus:      "protective_signal",
			candidateResponse: "prefer_recovery",
		},
		{
			name:              "neutral response",
			targetRPE:         6,
			input:             CompletionInput{ActualRPE: 6, Difficulty: "moderate", FatigueAfter: 3},
			periods:           validHistoryPeriods(),
			rulesV1Kind:       "",
			shadowStatus:      "observation_only",
			candidateResponse: "maintain_observed",
		},
		{
			name:      "recent recovery need",
			targetRPE: 6,
			input:     CompletionInput{ActualRPE: 6, Difficulty: "moderate", FatigueAfter: 3},
			periods: func() []TrainingHistoryPeriod {
				periods := validHistoryPeriods()
				periods[0].RecoveryNeededCheckins = 1
				return periods
			}(),
			rulesV1Kind:       "",
			shadowStatus:      "protective_signal",
			candidateResponse: "prefer_recovery",
		},
		{
			name:              "easy response without evidence",
			targetRPE:         6,
			input:             CompletionInput{ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2},
			periods:           nil,
			rulesV1Kind:       "progression",
			shadowStatus:      "not_evaluated",
			candidateResponse: "defer_progression",
		},
		{
			name:              "easy response with complete evidence",
			targetRPE:         6,
			input:             CompletionInput{ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2},
			periods:           validHistoryPeriods(),
			rulesV1Kind:       "progression",
			shadowStatus:      "observation_only",
			candidateResponse: "progress_duration_5pct",
		},
		{
			name:      "easy response with inconsistent evidence",
			targetRPE: 6,
			input:     CompletionInput{ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2},
			periods: func() []TrainingHistoryPeriod {
				periods := validHistoryPeriods()
				periods[0].PeriodDays = 6
				return periods
			}(),
			rulesV1Kind:       "progression",
			shadowStatus:      "not_evaluated",
			candidateResponse: "not_evaluated",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			activeDecision := DecideAdaptation(tc.targetRPE, tc.input)
			if activeDecision.Kind != tc.rulesV1Kind {
				t.Fatalf("rules-v1 kind = %q, want %q: %#v", activeDecision.Kind, tc.rulesV1Kind, activeDecision)
			}

			shadow := assessRulesV2AdaptationShadow(tc.targetRPE, tc.input, tc.periods, time.Unix(0, 0))
			if shadow.Status != tc.shadowStatus || shadow.CandidateResponse != tc.candidateResponse {
				t.Fatalf("shadow = %s/%s, want %s/%s: %+v", shadow.Status, shadow.CandidateResponse, tc.shadowStatus, tc.candidateResponse, shadow)
			}
			if shadow.ProgressionEligible || shadow.Applied || shadow.UsedForPrescription {
				t.Fatalf("shadow became authoritative: %+v", shadow)
			}
		})
	}
}
