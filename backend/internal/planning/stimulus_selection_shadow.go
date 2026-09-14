package planning

import "time"

const (
	stimulusSelectionShadowVersion = "stimulus-selection-shadow-v1"
	stimulusSelectionShadowMode    = "shadow"
	stimulusSelectionShadowScope   = "plan_generation_only"
)

// StimulusSelectionShadowAssessment records whether the selected session
// families match the needs that can be inferred from the current context. It
// is deliberately descriptive: rules-v1 remains the only prescribing engine.
type StimulusSelectionShadowAssessment struct {
	Version             string            `json:"version"`
	Mode                string            `json:"mode"`
	Scope               string            `json:"scope"`
	AssessedAt          string            `json:"assessed_at"`
	Status              string            `json:"status"`
	CandidateResponse   string            `json:"candidate_response"`
	CandidateNeed       string            `json:"candidate_need"`
	ExpectedStimuli     []string          `json:"expected_stimuli"`
	SelectedStimuli     []string          `json:"selected_stimuli"`
	RulesEvaluated      []string          `json:"rules_evaluated"`
	Reasons             []ReadinessReason `json:"reasons"`
	MissingData         []string          `json:"missing_data"`
	DataIssues          []string          `json:"data_issues"`
	NotEvaluated        []string          `json:"not_evaluated"`
	ProgressionEligible bool              `json:"progression_eligible"`
	Applied             bool              `json:"applied"`
	UsedForPrescription bool              `json:"used_for_prescription"`
}

func assessStimulusSelectionShadow(input Context, workouts []Workout, now time.Time, restricted bool) StimulusSelectionShadowAssessment {
	result := StimulusSelectionShadowAssessment{
		Version:           stimulusSelectionShadowVersion,
		Mode:              stimulusSelectionShadowMode,
		Scope:             stimulusSelectionShadowScope,
		AssessedAt:        now.UTC().Format(time.RFC3339Nano),
		Status:            "not_evaluated",
		CandidateResponse: "defer_evaluation",
		RulesEvaluated: []string{
			"need_classification_gate",
			"safety_precedence_gate",
			"base_and_adherence_gate",
			"event_specificity_gate",
			"stimulus_match_gate",
			"prescription_isolation_gate",
		},
		Reasons:             []ReadinessReason{},
		MissingData:         []string{},
		DataIssues:          []string{},
		NotEvaluated:        []string{"prescription_effect", "longitudinal_effect", "confidence_calibration", "activities_outside_cadencia", "athlete_timezone"},
		ExpectedStimuli:     []string{},
		SelectedStimuli:     []string{},
		ProgressionEligible: false,
		Applied:             false,
		UsedForPrescription: false,
	}

	addReason := func(code, message string) {
		for _, reason := range result.Reasons {
			if reason.Code == code {
				return
			}
		}
		result.Reasons = append(result.Reasons, ReadinessReason{Code: code, Message: message})
	}
	addMissing := func(value string) {
		result.MissingData = appendUniqueString(result.MissingData, value)
	}
	addIssue := func(value string) {
		result.DataIssues = appendUniqueString(result.DataIssues, value)
	}

	need, expected, missing := classifyStimulusNeed(input, restricted)
	result.CandidateNeed = need
	result.ExpectedStimuli = append(result.ExpectedStimuli, expected...)
	for _, value := range missing {
		addMissing(value)
	}
	for _, workout := range workouts {
		protocolKey := workoutProtocolKey(workout)
		if protocolKey == "" {
			addIssue("missing_workout_protocol")
			continue
		}
		result.SelectedStimuli = appendUniqueString(result.SelectedStimuli, protocolKey)
		if workout.DurationMinutes <= 0 {
			addIssue("invalid_planned_duration")
		}
		if workout.TargetRPE < 1 || workout.TargetRPE > 10 {
			addIssue("invalid_planned_target_rpe")
		}
	}
	if len(workouts) == 0 {
		addMissing("planned_workouts")
	}
	if len(result.SelectedStimuli) == 0 && len(workouts) > 0 {
		addMissing("classified_stimuli")
	}

	if len(result.DataIssues) > 0 {
		addReason("inconsistent_selection_data", "Há dados inválidos na sessão planejada; a coerência entre necessidade e estímulo permanece não avaliada.")
		return result
	}
	if len(result.MissingData) > 0 {
		addReason("insufficient_selection_context", "O contexto ou as sessões planejadas não têm cobertura suficiente para avaliar a escolha do estímulo.")
		return result
	}

	if mismatches := stimulusSelectionMismatches(result.CandidateNeed, workouts); len(mismatches) > 0 {
		result.DataIssues = append(result.DataIssues, mismatches...)
		addReason("stimulus_need_mismatch", "A seleção observada não coincide completamente com a necessidade inferida; isso é um sinal para revisão, não uma alteração automática do plano.")
		result.Status = "observed_mismatch"
		result.CandidateResponse = "review_selection"
		return result
	}

	addReason("stimulus_need_aligned", "A família de estímulos selecionada é compatível com a necessidade inferida a partir do contexto disponível.")
	addReason("shadow_only", "A leitura não altera os treinos; o rules-v1 continua responsável pela prescrição.")
	result.Status = "observed"
	result.CandidateResponse = "maintain_observed"
	return result
}

func classifyStimulusNeed(input Context, restricted bool) (string, []string, []string) {
	if restricted || input.Observed.RequiresRecovery() {
		return "recovery_protection", []string{"active_recovery", "protected_recovery"}, nil
	}
	if trainingHistorySuggestsLowAdherence(input.TrainingHistory) {
		return "base_and_adherence", []string{"base_endurance", "continuous_endurance", "long_endurance", "active_recovery"}, nil
	}
	if input.Cycling.EventGoal {
		if input.Cycling.EventDistanceKM == nil {
			return "event_specificity", []string{"continuous_endurance", "long_endurance", "controlled_event_pace"}, []string{"event_distance_km"}
		}
		if *input.Cycling.EventDistanceKM >= 150 {
			return "endurance_specificity", []string{"continuous_endurance", "long_endurance"}, nil
		}
		if *input.Cycling.EventDistanceKM <= 80 && input.ExperienceLevel == "advanced" && input.BaselineEligible {
			return "quality_progression", []string{"controlled_tempo", "controlled_intervals", "road_moderate_intervals", "road_vo2_intervals"}, nil
		}
		return "event_specificity", []string{"continuous_endurance", "long_endurance", "controlled_event_pace"}, nil
	}
	if input.PrimaryGoal == "performance" && input.ExperienceLevel == "advanced" && input.BaselineEligible {
		return "quality_progression", []string{"controlled_tempo", "controlled_intervals", "road_moderate_intervals", "road_vo2_intervals", "long_endurance"}, nil
	}
	return "general_base", []string{"base_endurance", "continuous_endurance", "long_endurance", "controlled_tempo"}, nil
}

func trainingHistorySuggestsLowAdherence(history []TrainingHistoryWindow) bool {
	if len(history) == 0 {
		return true
	}
	for _, window := range history {
		if window.WindowDays == 28 {
			return window.MissedSessions > 0 || window.OverdueInProgressSessions > 0 || window.ExpectedSessions == 0
		}
	}
	return true
}

func stimulusSelectionMismatches(need string, workouts []Workout) []string {
	quality := false
	highIntensity := false
	recovery := false
	longEndurance := false
	for _, workout := range workouts {
		key := workoutProtocolKey(workout)
		if workout.TargetRPE >= qualityTargetRPEThreshold {
			quality = true
		}
		if workout.TargetRPE >= highIntensityRPEThreshold {
			highIntensity = true
		}
		if key == "active_recovery" || key == "protected_recovery" || workout.TargetRPE <= 3.5 {
			recovery = true
		}
		if key == "continuous_endurance" || key == "long_endurance" {
			longEndurance = true
		}
	}

	var mismatches []string
	switch need {
	case "recovery_protection":
		if highIntensity {
			mismatches = append(mismatches, "quality_in_recovery_need")
		}
		if !recovery {
			mismatches = append(mismatches, "missing_recovery_stimulus")
		}
	case "base_and_adherence":
		if highIntensity {
			mismatches = append(mismatches, "high_intensity_with_low_adherence")
		}
	case "endurance_specificity":
		if !longEndurance {
			mismatches = append(mismatches, "missing_long_endurance")
		}
	case "quality_progression":
		if !quality {
			mismatches = append(mismatches, "missing_quality_stimulus")
		}
	}
	return mismatches
}
