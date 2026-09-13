package planning

import "time"

const (
	planningCoherenceShadowVersion = "planning-coherence-shadow-v1"
	planningCoherenceShadowMode    = "shadow"
	planningCoherenceShadowScope   = "plan_generation_only"
)

// PlanningCoherenceComponentObservation records the contract state of one
// shadow component without copying its full assessment into the aggregate.
type PlanningCoherenceComponentObservation struct {
	Component         string `json:"component"`
	Version           string `json:"version"`
	Mode              string `json:"mode"`
	Status            string `json:"status"`
	CandidateResponse string `json:"candidate_response,omitempty"`
}

// PlanningCoherenceShadowAssessment checks whether periodization, observed
// stimulus distribution and selected stimulus families tell a compatible
// story. It is descriptive only: rules-v1 remains the prescribing engine.
type PlanningCoherenceShadowAssessment struct {
	Version                                        string                                  `json:"version"`
	Mode                                           string                                  `json:"mode"`
	Scope                                          string                                  `json:"scope"`
	AssessedAt                                     string                                  `json:"assessed_at"`
	Status                                         string                                  `json:"status"`
	CandidateResponse                              string                                  `json:"candidate_response"`
	Components                                     []PlanningCoherenceComponentObservation `json:"components"`
	CandidateNeed                                  string                                  `json:"candidate_need"`
	CoherenceChecks                                []string                                `json:"coherence_checks"`
	PeriodizationRecoveryWeekQualitySessions       int                                     `json:"periodization_recovery_week_quality_sessions"`
	PeriodizationRecoveryWeekHighIntensitySessions int                                     `json:"periodization_recovery_week_high_intensity_sessions"`
	DistributionQualitySessionsLast7d              int                                     `json:"distribution_quality_sessions_last_7d"`
	DistributionAdjacentQualitySessionPairs        int                                     `json:"distribution_adjacent_quality_session_pairs"`
	RulesEvaluated                                 []string                                `json:"rules_evaluated"`
	Reasons                                        []ReadinessReason                       `json:"reasons"`
	MissingData                                    []string                                `json:"missing_data"`
	DataIssues                                     []string                                `json:"data_issues"`
	NotEvaluated                                   []string                                `json:"not_evaluated"`
	ProgressionEligible                            bool                                    `json:"progression_eligible"`
	Applied                                        bool                                    `json:"applied"`
	UsedForPrescription                            bool                                    `json:"used_for_prescription"`
}

func assessPlanningCoherenceShadow(periodization PeriodizationShadowAssessment, distribution TrainingStimulusDistribution, selection StimulusSelectionShadowAssessment, now time.Time) PlanningCoherenceShadowAssessment {
	result := PlanningCoherenceShadowAssessment{
		Version:           planningCoherenceShadowVersion,
		Mode:              planningCoherenceShadowMode,
		Scope:             planningCoherenceShadowScope,
		AssessedAt:        now.UTC().Format(time.RFC3339Nano),
		Status:            "not_evaluated",
		CandidateResponse: "defer_evaluation",
		Components: []PlanningCoherenceComponentObservation{
			{
				Component:         "periodization",
				Version:           periodization.Version,
				Mode:              periodization.Mode,
				Status:            periodization.Status,
				CandidateResponse: periodization.CandidateResponse,
			},
			{
				Component: "stimulus_distribution",
				Version:   distribution.Version,
				Mode:      distribution.Mode,
				Status:    stimulusDistributionObservationStatus(distribution),
			},
			{
				Component:         "stimulus_selection",
				Version:           selection.Version,
				Mode:              selection.Mode,
				Status:            selection.Status,
				CandidateResponse: selection.CandidateResponse,
			},
		},
		CandidateNeed:                                  selection.CandidateNeed,
		CoherenceChecks:                                []string{},
		PeriodizationRecoveryWeekQualitySessions:       recoveryWeekQualitySessions(periodization),
		PeriodizationRecoveryWeekHighIntensitySessions: recoveryWeekHighIntensitySessions(periodization),
		DistributionQualitySessionsLast7d:              distribution.QualitySessionsLast7d,
		DistributionAdjacentQualitySessionPairs:        distribution.AdjacentQualitySessionPairs,
		RulesEvaluated: []string{
			"periodization_presence_gate",
			"stimulus_distribution_presence_gate",
			"stimulus_selection_presence_gate",
			"component_contract_gate",
			"recovery_precedence_coherence_gate",
			"phase_stimulus_coherence_gate",
			"distribution_selection_coherence_gate",
			"prescription_isolation_gate",
		},
		Reasons:             []ReadinessReason{},
		MissingData:         []string{},
		DataIssues:          []string{},
		NotEvaluated:        []string{"prescription_effect", "longitudinal_effect", "confidence_calibration", "athlete_timezone"},
		ProgressionEligible: false,
		Applied:             false,
		UsedForPrescription: false,
	}

	addMissing := func(value string) {
		result.MissingData = appendUniqueString(result.MissingData, value)
	}
	addIssue := func(value string) {
		result.DataIssues = appendUniqueString(result.DataIssues, value)
	}
	addCheck := func(value string) {
		result.CoherenceChecks = appendUniqueString(result.CoherenceChecks, value)
	}
	addReason := func(code, message string) {
		for _, reason := range result.Reasons {
			if reason.Code == code {
				return
			}
		}
		result.Reasons = append(result.Reasons, ReadinessReason{Code: code, Message: message})
	}

	if periodization.Version == "" {
		addMissing("periodization_shadow")
	}
	if distribution.Version == "" {
		addMissing("stimulus_distribution")
	}
	if selection.Version == "" {
		addMissing("stimulus_selection_shadow")
	}

	if periodization.Mode != "" && periodization.Mode != periodizationShadowMode {
		addIssue("periodization_shadow_mode")
	}
	if distribution.Mode != "" && distribution.Mode != trainingHistoryMode {
		addIssue("stimulus_distribution_mode")
	}
	if selection.Mode != "" && selection.Mode != stimulusSelectionShadowMode {
		addIssue("stimulus_selection_shadow_mode")
	}
	if periodization.ProgressionEligible || periodization.Applied || periodization.UsedForPrescription ||
		distribution.UsedForPrescription || selection.ProgressionEligible || selection.Applied || selection.UsedForPrescription {
		addIssue("component_authority_violation")
	}

	if len(periodization.DataIssues) > 0 {
		addIssue("periodization_shadow_inconsistent")
	}
	if periodization.Status != "observed" {
		addMissing("periodization_observation")
	}
	if stimulusDistributionHasIncompleteData(distribution) {
		for _, value := range distribution.MissingData {
			addMissing("stimulus_distribution_" + value)
		}
		for _, value := range distribution.DataIssues {
			addIssue("stimulus_distribution_" + value)
		}
	}
	if selection.Status == "not_evaluated" {
		addMissing("stimulus_selection_observation")
	}
	if selection.Status == "observed_mismatch" {
		addIssue("stimulus_selection_mismatch")
	}

	if selection.Status != "not_evaluated" && periodization.Status == "observed" && !stimulusDistributionHasIncompleteData(distribution) {
		assessPlanningCoherenceSignals(&result, periodization, distribution, selection, addCheck, addIssue)
	}

	if len(result.DataIssues) > 0 {
		addReason("coherence_mismatch", "As auditorias de periodização, distribuição e seleção apresentam uma divergência observável; isso requer revisão e não altera o plano.")
		result.Status = "observed_mismatch"
		result.CandidateResponse = "review_coherence"
		return result
	}
	if len(result.MissingData) > 0 {
		addReason("insufficient_coherence_data", "Os três componentes ainda não têm cobertura suficiente para concluir a coerência do planejamento.")
		return result
	}

	addReason("planning_coherence_observed", "Periodização, distribuição histórica e seleção de estímulos permaneceram coerentes no recorte observado.")
	addReason("shadow_only", "A auditoria é somente observacional; o rules-v1 continua responsável pela prescrição.")
	result.Status = "observed"
	result.CandidateResponse = "maintain_observed"
	return result
}

func assessPlanningCoherenceSignals(result *PlanningCoherenceShadowAssessment, periodization PeriodizationShadowAssessment, distribution TrainingStimulusDistribution, selection StimulusSelectionShadowAssessment, addCheck func(string), addIssue func(string)) {
	if selection.CandidateNeed == "recovery_protection" {
		if result.PeriodizationRecoveryWeekQualitySessions > 0 || result.PeriodizationRecoveryWeekHighIntensitySessions > 0 {
			addIssue("recovery_need_not_reflected_in_periodization")
		} else if len(periodization.Weeks) == periodizationWeekCount && periodization.Weeks[periodizationWeekCount-1].RecoverySessions > 0 {
			addCheck("recovery_need_matches_recovery_phase")
		} else {
			addIssue("recovery_need_without_recovery_phase")
		}
	}

	if stimulusDistributionRequiresRecovery(distribution) {
		if selection.CandidateNeed != "recovery_protection" {
			addIssue("distribution_protection_not_reflected_in_selection")
		} else {
			addCheck("distribution_protection_matches_selection")
		}
	}

	switch selection.CandidateNeed {
	case "endurance_specificity":
		if periodizationLongSessions(periodization) == 0 {
			addIssue("endurance_need_without_long_session")
		} else {
			addCheck("endurance_need_matches_periodization")
		}
	case "quality_progression":
		if periodization.QualitySessions == 0 {
			addIssue("quality_need_without_quality_session")
		} else {
			addCheck("quality_need_matches_periodization")
		}
	}
}

func stimulusDistributionObservationStatus(distribution TrainingStimulusDistribution) string {
	if distribution.Version == "" || stimulusDistributionHasIncompleteData(distribution) {
		return "not_evaluated"
	}
	return "observed"
}

func recoveryWeekQualitySessions(assessment PeriodizationShadowAssessment) int {
	if len(assessment.Weeks) != periodizationWeekCount {
		return 0
	}
	return assessment.Weeks[periodizationWeekCount-1].QualitySessions
}

func recoveryWeekHighIntensitySessions(assessment PeriodizationShadowAssessment) int {
	if len(assessment.Weeks) != periodizationWeekCount {
		return 0
	}
	return assessment.Weeks[periodizationWeekCount-1].HighIntensitySessions
}

func periodizationLongSessions(assessment PeriodizationShadowAssessment) int {
	count := 0
	for _, week := range assessment.Weeks {
		count += week.LongSessions
	}
	return count
}
