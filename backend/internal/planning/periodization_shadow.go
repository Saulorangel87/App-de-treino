package planning

import (
	"fmt"
	"sort"
	"time"
)

const (
	periodizationShadowVersion = "periodization-shadow-v1"
	periodizationShadowMode    = "shadow"
	periodizationShadowScope   = "plan_generation_only"
	periodizationWeekCount     = 4
)

// PeriodizationWeekObservation describes the planned shape of one week. It
// deliberately observes the generated plan instead of deciding its workouts.
type PeriodizationWeekObservation struct {
	WeekIndex                         int      `json:"week_index"`
	Phase                             string   `json:"phase"`
	PlannedSessions                   int      `json:"planned_sessions"`
	TotalPlannedMinutes               int      `json:"total_planned_minutes"`
	QualitySessions                   int      `json:"quality_sessions"`
	HighIntensitySessions             int      `json:"high_intensity_sessions"`
	RecoverySessions                  int      `json:"recovery_sessions"`
	LongSessions                      int      `json:"long_sessions"`
	TaperedSessions                   int      `json:"tapered_sessions"`
	AverageTargetRPE                  *float64 `json:"average_target_rpe"`
	MinimumDaysBetweenQualitySessions *int     `json:"minimum_days_between_quality_sessions"`
}

// PeriodizationShadowAssessment records whether the generated cycle has the
// intended broad structure. It is not a second prescription engine: rules-v1
// remains responsible for generating the plan.
type PeriodizationShadowAssessment struct {
	Version                           string                         `json:"version"`
	Mode                              string                         `json:"mode"`
	Scope                             string                         `json:"scope"`
	AssessedAt                        string                         `json:"assessed_at"`
	Status                            string                         `json:"status"`
	CandidateResponse                 string                         `json:"candidate_response"`
	RulesEvaluated                    []string                       `json:"rules_evaluated"`
	Reasons                           []ReadinessReason              `json:"reasons"`
	MissingData                       []string                       `json:"missing_data"`
	DataIssues                        []string                       `json:"data_issues"`
	NotEvaluated                      []string                       `json:"not_evaluated"`
	Weeks                             []PeriodizationWeekObservation `json:"weeks"`
	TotalPlannedMinutes               int                            `json:"total_planned_minutes"`
	QualitySessions                   int                            `json:"quality_sessions"`
	HighIntensitySessions             int                            `json:"high_intensity_sessions"`
	AdjacentQualitySessionPairs       int                            `json:"adjacent_quality_session_pairs"`
	MinimumDaysBetweenQualitySessions *int                           `json:"minimum_days_between_quality_sessions"`
	ProgressionEligible               bool                           `json:"progression_eligible"`
	Applied                           bool                           `json:"applied"`
	UsedForPrescription               bool                           `json:"used_for_prescription"`
}

type periodizationQualityDate struct {
	Day       time.Time
	WeekIndex int
}

func assessPeriodizationShadow(workouts []Workout, start, now time.Time) PeriodizationShadowAssessment {
	result := PeriodizationShadowAssessment{
		Version:           periodizationShadowVersion,
		Mode:              periodizationShadowMode,
		Scope:             periodizationShadowScope,
		AssessedAt:        now.UTC().Format(time.RFC3339Nano),
		Status:            "not_evaluated",
		CandidateResponse: "defer_evaluation",
		RulesEvaluated: []string{
			"four_week_structure_gate",
			"recovery_week_gate",
			"quality_density_gate",
			"quality_spacing_gate",
			"volume_progression_gate",
			"prescription_isolation_gate",
		},
		Reasons:     []ReadinessReason{},
		MissingData: []string{},
		DataIssues:  []string{},
		NotEvaluated: []string{
			"load_tolerance",
			"detraining",
			"fitness_change",
			"activities_outside_cadencia",
			"athlete_timezone",
			"prescription_effect",
		},
		Weeks: make([]PeriodizationWeekObservation, periodizationWeekCount),
	}
	for index := range result.Weeks {
		phase := "progression"
		if index == periodizationWeekCount-1 {
			phase = "recovery"
		}
		result.Weeks[index] = PeriodizationWeekObservation{WeekIndex: index, Phase: phase}
	}

	addMissing := func(value string) {
		result.MissingData = appendUniqueString(result.MissingData, value)
	}
	addIssue := func(value string) {
		result.DataIssues = appendUniqueString(result.DataIssues, value)
	}
	addReason := func(code, message string) {
		for _, reason := range result.Reasons {
			if reason.Code == code {
				return
			}
		}
		result.Reasons = append(result.Reasons, ReadinessReason{Code: code, Message: message})
	}

	startDay := dateOnly(start)
	qualityDates := make([]periodizationQualityDate, 0)
	for _, workout := range workouts {
		if workout.DurationMinutes <= 0 {
			addIssue("invalid_planned_duration")
		}
		if workout.TargetRPE < 1 || workout.TargetRPE > 10 {
			addIssue("invalid_planned_target_rpe")
		}
		day, err := time.Parse("2006-01-02", workout.ScheduledOn)
		if err != nil {
			addIssue("invalid_scheduled_date")
			continue
		}
		day = dateOnly(day)
		daysFromStart := int(day.Sub(startDay).Hours() / 24)
		if daysFromStart < 0 || daysFromStart >= periodizationWeekCount*7 {
			addIssue("scheduled_date_outside_four_week_cycle")
			continue
		}
		weekIndex := daysFromStart / 7
		week := &result.Weeks[weekIndex]
		week.PlannedSessions++
		week.TotalPlannedMinutes += workout.DurationMinutes
		result.TotalPlannedMinutes += workout.DurationMinutes
		if workout.TargetRPE >= qualityTargetRPEThreshold {
			week.QualitySessions++
			result.QualitySessions++
			qualityDates = append(qualityDates, periodizationQualityDate{Day: day, WeekIndex: weekIndex})
		}
		if workout.TargetRPE >= highIntensityRPEThreshold {
			week.HighIntensitySessions++
			result.HighIntensitySessions++
		}
		protocolKey := workoutProtocolKey(workout)
		if protocolKey == "active_recovery" || protocolKey == "protected_recovery" || workout.TargetRPE <= 3.5 {
			week.RecoverySessions++
		}
		if protocolKey == "continuous_endurance" || protocolKey == "long_endurance" {
			week.LongSessions++
		}
		if explanationBool(workout.Explanation, "event_taper_applied") {
			week.TaperedSessions++
		}
	}

	for index := range result.Weeks {
		week := &result.Weeks[index]
		if week.PlannedSessions == 0 {
			addMissing(fmt.Sprintf("week_%d_sessions", index+1))
			continue
		}
		average := float64(week.TotalPlannedMinutes)
		average /= float64(week.PlannedSessions)
		week.AverageTargetRPE = &average
		if week.QualitySessions > 1 {
			addIssue(fmt.Sprintf("multiple_quality_sessions_week_%d", index+1))
		}
	}

	sort.Slice(qualityDates, func(left, right int) bool {
		return qualityDates[left].Day.Before(qualityDates[right].Day)
	})
	for index := 1; index < len(qualityDates); index++ {
		gap := int(qualityDates[index].Day.Sub(qualityDates[index-1].Day).Hours() / 24)
		if gap < 0 {
			addIssue("quality_dates_out_of_order")
			continue
		}
		if result.MinimumDaysBetweenQualitySessions == nil || gap < *result.MinimumDaysBetweenQualitySessions {
			value := gap
			result.MinimumDaysBetweenQualitySessions = &value
		}
		previousWeek := &result.Weeks[qualityDates[index-1].WeekIndex]
		currentWeek := &result.Weeks[qualityDates[index].WeekIndex]
		if previousWeek.MinimumDaysBetweenQualitySessions == nil || gap < *previousWeek.MinimumDaysBetweenQualitySessions {
			value := gap
			previousWeek.MinimumDaysBetweenQualitySessions = &value
		}
		if currentWeek.MinimumDaysBetweenQualitySessions == nil || gap < *currentWeek.MinimumDaysBetweenQualitySessions {
			value := gap
			currentWeek.MinimumDaysBetweenQualitySessions = &value
		}
		if gap <= 1 {
			result.AdjacentQualitySessionPairs++
			addIssue("quality_sessions_on_adjacent_days")
		}
	}

	recoveryWeek := result.Weeks[periodizationWeekCount-1]
	if recoveryWeek.QualitySessions > 0 || recoveryWeek.HighIntensitySessions > 0 {
		addIssue("recovery_week_contains_quality")
	}
	if previousWeek := result.Weeks[periodizationWeekCount-2]; previousWeek.PlannedSessions > 0 && recoveryWeek.PlannedSessions > 0 && recoveryWeek.TotalPlannedMinutes >= previousWeek.TotalPlannedMinutes {
		addIssue("recovery_week_not_lower_volume")
	}

	if len(result.DataIssues) > 0 {
		addReason("inconsistent_periodization", "A estrutura planejada contém uma inconsistência; a leitura observacional não conclui se o ciclo é coerente.")
		return result
	}
	if len(result.MissingData) > 0 {
		addReason("insufficient_periodization_data", "O ciclo não tem sessões planejadas em todas as quatro semanas; a periodização permanece não avaliada.")
		return result
	}

	addReason("four_week_periodization_observed", "O ciclo observado preserva três semanas de progressão e uma semana de recuperação, com volume e estímulos registrados apenas para auditoria.")
	result.Status = "observed"
	result.CandidateResponse = "maintain_observed"
	return result
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func workoutProtocolKey(workout Workout) string {
	if value, ok := workout.Explanation["protocol_key"].(string); ok {
		return value
	}
	if value, ok := workout.Structure["protocol_key"].(string); ok {
		return value
	}
	return ""
}

func explanationBool(explanation map[string]any, key string) bool {
	value, ok := explanation[key].(bool)
	return ok && value
}
