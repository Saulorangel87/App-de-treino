package planning

import "time"

const (
	postEventRecoveryVersion    = "post-event-recovery-v1"
	postEventRecoveryScope      = "post_event_plan_load"
	postEventRecoveryWindowDays = 7
	postEventRecoveryMaxMinutes = 45
	postEventRecoveryTargetRPE  = 3.5
)

var postEventRecoveryEvidenceKeys = []string{"post-competition-recovery-2019", "recovery-umbrella-2024"}

// PostEventRecoveryAssessment keeps the short post-event window explicit. It
// is a conservative load reduction, not a claim that one recovery protocol is
// universally effective for every cyclist.
type PostEventRecoveryAssessment struct {
	Version             string            `json:"version"`
	Mode                string            `json:"mode"`
	Scope               string            `json:"scope"`
	AssessedAt          string            `json:"assessed_at"`
	Status              string            `json:"status"`
	Applied             bool              `json:"applied"`
	UsedForPrescription bool              `json:"used_for_prescription"`
	EventDate           *string           `json:"event_date,omitempty"`
	DaysAfterEvent      *int              `json:"days_after_event,omitempty"`
	MaxDurationMinutes  int               `json:"max_duration_minutes"`
	TargetRPE           float64           `json:"target_rpe"`
	EvidenceKeys        []string          `json:"evidence_keys"`
	RulesEvaluated      []string          `json:"rules_evaluated"`
	Reasons             []ReadinessReason `json:"reasons"`
}

func assessPostEventRecovery(input Context, now time.Time) PostEventRecoveryAssessment {
	result := PostEventRecoveryAssessment{
		Version:            postEventRecoveryVersion,
		Mode:               "prescriptive",
		Scope:              postEventRecoveryScope,
		AssessedAt:         now.UTC().Format(time.RFC3339Nano),
		Status:             "not_applicable",
		MaxDurationMinutes: postEventRecoveryMaxMinutes,
		TargetRPE:          postEventRecoveryTargetRPE,
		EvidenceKeys:       append([]string(nil), postEventRecoveryEvidenceKeys...),
		RulesEvaluated: []string{
			"event_goal_gate",
			"event_date_gate",
			"post_event_window_gate",
			"protective_signal_precedence_gate",
		},
		Reasons: []ReadinessReason{},
	}
	addReason := func(code, message string) {
		result.Reasons = append(result.Reasons, ReadinessReason{Code: code, Message: message})
	}
	if !input.Cycling.EventGoal {
		addReason("event_goal_not_declared", "A preparação para um evento não foi informada.")
		return result
	}
	daysUntil, ok := daysUntilEvent(input.Cycling, now)
	if !ok || daysUntil >= 0 {
		addReason("event_not_recently_finished", "Não há evento encerrado recentemente para ativar a janela de recuperação pós-prova.")
		return result
	}
	daysAfter := -daysUntil
	result.EventDate = input.Cycling.EventDate
	result.DaysAfterEvent = &daysAfter
	if daysAfter > postEventRecoveryWindowDays {
		addReason("post_event_window_expired", "A janela conservadora de até sete dias após o evento já terminou.")
		return result
	}
	if len(input.Limitations) > 0 || isReturningAfterPause(input.Cycling) || input.Observed.RequiresRecovery() {
		result.Status = "protective_signal"
		addReason("existing_protection_precedes_event_template", "Uma proteção ativa de limitação, retorno ou recuperação insuficiente continua prioritária; a recuperação pós-prova não altera a prescrição.")
		return result
	}
	result.Status = "eligible"
	result.Applied = true
	result.UsedForPrescription = true
	addReason("eligible_for_post_event_recovery", "O evento terminou há poucos dias; o plano reduz temporariamente duração e esforço antes de retomar estímulos maiores.")
	return result
}

func postEventRecoveryAppliesToWorkout(cycling CyclingContext, assessment PostEventRecoveryAssessment, date time.Time) bool {
	if !assessment.Applied {
		return false
	}
	daysUntil, ok := daysUntilEvent(cycling, date)
	return ok && daysUntil < 0 && daysUntil >= -postEventRecoveryWindowDays
}
