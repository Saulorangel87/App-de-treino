package planning

import "time"

const (
	eventTaperVersion           = "taper-v1"
	eventTaperScope             = "event_based_plan_volume"
	eventTaperMinDaysFromToday  = 7
	eventTaperMaxDaysFromToday  = 21
	eventTaperSessionWindowDays = 14
	eventTaperVolumeMultiplier  = 0.5
)

var eventTaperEvidenceKeys = []string{"taper-cyclist-2025", "taper-meta-2023"}

// EventTaperAssessment records the explicit gates used by the event taper.
// It is prescriptive only when all conservative eligibility conditions pass.
type EventTaperAssessment struct {
	Version             string            `json:"version"`
	Mode                string            `json:"mode"`
	Scope               string            `json:"scope"`
	AssessedAt          string            `json:"assessed_at"`
	Status              string            `json:"status"`
	Applied             bool              `json:"applied"`
	UsedForPrescription bool              `json:"used_for_prescription"`
	EventDate           *string           `json:"event_date,omitempty"`
	DaysFromToday       *int              `json:"days_from_today,omitempty"`
	VolumeMultiplier    float64           `json:"volume_multiplier"`
	EvidenceKeys        []string          `json:"evidence_keys"`
	RulesEvaluated      []string          `json:"rules_evaluated"`
	Reasons             []ReadinessReason `json:"reasons"`
}

func assessEventTaper(input Context, now time.Time, restricted bool) EventTaperAssessment {
	result := EventTaperAssessment{
		Version:          eventTaperVersion,
		Mode:             "prescriptive",
		Scope:            eventTaperScope,
		AssessedAt:       now.UTC().Format(time.RFC3339Nano),
		Status:           "not_applicable",
		VolumeMultiplier: 1,
		EvidenceKeys:     append([]string(nil), eventTaperEvidenceKeys...),
		RulesEvaluated: []string{
			"event_goal_gate",
			"event_date_gate",
			"training_base_gate",
			"baseline_gate",
			"protective_signal_gate",
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
	eventDays, ok := daysUntilEvent(input.Cycling, now)
	if !ok {
		addReason("invalid_event_date", "A data do evento está ausente, inválida ou não é futura.")
		return result
	}
	result.EventDate = input.Cycling.EventDate
	result.DaysFromToday = &eventDays
	if eventDays < 0 {
		addReason("invalid_event_date", "A data do evento está ausente, inválida ou não é futura.")
		return result
	}
	if eventDays < eventTaperMinDaysFromToday || eventDays > eventTaperMaxDaysFromToday {
		addReason("event_outside_taper_window", "O evento não está na janela de 7 a 21 dias usada por este piloto.")
		return result
	}
	if input.ExperienceLevel != "advanced" {
		addReason("advanced_level_required", "A evidência disponível foi estudada em ciclistas treinados; o piloto exige nível avançado.")
		return result
	}
	if !input.BaselineEligible {
		addReason("baseline_not_eligible", "A avaliação submáxima apta é necessária para este piloto.")
		return result
	}
	if input.Cycling.RecentTrainingWeeks < 8 || input.Cycling.WeeklyRides < 3 {
		addReason("training_base_insufficient", "São necessárias pelo menos 8 semanas de treino e 3 pedais semanais informados.")
		return result
	}
	if restricted {
		addReason("active_limitation", "Uma limitação ativa mantém as proteções existentes prioritárias.")
		return result
	}
	if input.Observed.RequiresRecovery() {
		addReason("recovery_needed", "Sinais recentes de dor ou recuperação insuficiente bloqueiam o taper específico.")
		return result
	}

	result.Status = "eligible"
	result.Applied = true
	result.UsedForPrescription = true
	result.VolumeMultiplier = eventTaperVolumeMultiplier
	addReason("eligible_for_event_taper", "O evento está na janela do piloto e os requisitos de treino, avaliação e segurança foram atendidos.")
	return result
}

func eventTaperAppliesToWorkout(input CyclingContext, assessment EventTaperAssessment, date time.Time, recoveryWeek bool) bool {
	if !assessment.Applied || recoveryWeek {
		return false
	}
	days, ok := daysUntilEvent(input, date)
	return ok && days > 0 && days <= eventTaperSessionWindowDays
}

func daysUntilEvent(cycling CyclingContext, value time.Time) (int, bool) {
	if !cycling.EventGoal || cycling.EventDate == nil || *cycling.EventDate == "" {
		return 0, false
	}
	eventDate, err := time.ParseInLocation("2006-01-02", *cycling.EventDate, value.Location())
	if err != nil {
		return 0, false
	}
	day := time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
	eventDay := time.Date(eventDate.Year(), eventDate.Month(), eventDate.Day(), 0, 0, 0, 0, value.Location())
	return int(eventDay.Sub(day).Hours() / 24), true
}
