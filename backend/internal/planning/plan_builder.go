package planning

import (
	"fmt"
	"sort"
	"time"
)

func buildPlan(input Context, now time.Time) (Plan, error) {
	if input.ProfileID == "" || input.PrimaryGoal == "" || len(input.Availability) == 0 {
		return Plan{}, ErrIncompleteOnboarding
	}
	maxSessions := map[string]int{"beginner": 3, "intermediate": 4, "advanced": 5}[input.ExperienceLevel]
	if maxSessions == 0 {
		return Plan{}, ErrIncompleteOnboarding
	}

	lowCurrentActivity := isLowCurrentActivity(input.Profile.ActivityLevel)
	slots := append([]AvailabilitySlot(nil), input.Availability...)
	sort.Slice(slots, func(i, j int) bool { return slots[i].AvailableMinutes > slots[j].AvailableMinutes })
	if len(slots) > maxSessions {
		slots = slots[:maxSessions]
	}
	sort.Slice(slots, func(i, j int) bool { return slots[i].Weekday < slots[j].Weekday })

	start := nextMonday(now)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	restricted := len(input.Limitations) > 0
	medicalRestriction := false
	for _, item := range input.Limitations {
		if item.ProfessionalClearanceRecommended {
			restricted = true
		}
		if item.MedicalRestriction {
			medicalRestriction = true
		}
	}
	workouts := make([]Workout, 0, len(slots)*4)
	multipliers := []float64{0.85, 0.95, 1.0, 0.75}
	eventTaper := assessEventTaper(input, now, restricted)
	postEventRecovery := assessPostEventRecovery(input, now)
	lowObservedAdherence := hasLowObservedAdherence(input.TrainingHistory)
	for week := 0; week < 4; week++ {
		longIndex := longestSlot(slots)
		intensityIndex := intensitySlot(slots, longIndex)
		recoveryWeek := week == len(multipliers)-1
		for index, slot := range slots {
			scheduledOn := start.AddDate(0, 0, week*7+weekdayOffset(slot.Weekday))
			if week == 0 && scheduledOn.Before(today) {
				continue
			}
			kind := "base"
			if index == longIndex {
				kind = "long"
			} else if index == intensityIndex && !restricted && !recoveryWeek && !lowObservedAdherence && !lowCurrentActivity {
				kind = "quality"
			}
			workouts = append(workouts, makeWorkout(input, slot, kind, restricted, multipliers[week], week, scheduledOn, eventTaper, postEventRecovery))
		}
	}

	periodizationShadow := assessPeriodizationShadow(workouts, start, now)
	stimulusSelectionShadow := assessStimulusSelectionShadow(input, workouts, now, restricted)
	trainingHistory := buildTrainingHistorySnapshot(input.TrainingHistory, now, input.TrainingHistoryPeriods)
	planningCoherenceShadow := assessPlanningCoherenceShadow(periodizationShadow, trainingHistory.StimulusDistribution, stimulusSelectionShadow, now)
	plan := Plan{
		StartsOn: start.Format("2006-01-02"),
		EndsOn:   start.AddDate(0, 0, 27).Format("2006-01-02"),
		Status:   "draft",
		PrescriptionSnapshot: map[string]any{
			"engine_version":            "rules-v1",
			"event_taper":               eventTaper,
			"post_event_recovery":       postEventRecovery,
			"rules_v2_shadow":           assessRulesV2Shadow(input, now),
			"periodization_shadow":      periodizationShadow,
			"stimulus_selection_shadow": stimulusSelectionShadow,
			"planning_coherence_shadow": planningCoherenceShadow,
			"readiness_assessment":      assessReadiness(input, now),
			"training_history":          trainingHistory,
			"experience_level":          input.ExperienceLevel,
			"primary_goal":              input.PrimaryGoal,
			"secondary_goal":            input.SecondaryGoal,
			"profile_context": map[string]any{
				"birth_date":       input.Profile.BirthDate,
				"sex":              input.Profile.Sex,
				"height_cm":        input.Profile.HeightCM,
				"weight_kg":        input.Profile.WeightKG,
				"waist_cm":         input.Profile.WaistCM,
				"body_fat_percent": input.Profile.BodyFatPercent,
				"weight_trend":     input.Profile.WeightTrend,
				"activity_level":   input.Profile.ActivityLevel,
			},
			"restricted": restricted,
			"safety_context": map[string]any{
				"active_limitations":           len(input.Limitations),
				"medical_restriction":          medicalRestriction,
				"prescription_protected":       restricted,
				"recent_surgery":               hasRecentSurgery(input.Limitations),
				"exercise_prohibited":          hasExerciseProhibited(input.Limitations),
				"condition_affecting_exercise": hasConditionAffectingExercise(input.Limitations),
			},
			"sessions_per_week": len(slots),
			"cycling_context": map[string]any{
				"weekly_hours":              input.Cycling.WeeklyHours,
				"practice_duration_months":  input.Cycling.PracticeDurationMonths,
				"average_ride_minutes":      input.Cycling.AverageRideMinutes,
				"longest_ride_minutes":      input.Cycling.LongestRideMinutes,
				"weekly_rides":              input.Cycling.WeeklyRides,
				"recent_weekly_distance_km": input.Cycling.RecentWeeklyDistanceKM,
				"recent_training_weeks":     input.Cycling.RecentTrainingWeeks,
				"training_status":           input.Cycling.TrainingStatus,
				"recent_best_distance_km":   input.Cycling.RecentBestDistanceKM,
				"preferred_session_types":   input.Cycling.PreferredSessionTypes,
				"discipline":                input.Cycling.Discipline,
				"bike_type":                 input.Cycling.BikeType,
				"terrain":                   input.Cycling.Terrain,
				"uses_heart_rate":           input.Cycling.UsesHeartRate,
				"uses_power":                input.Cycling.UsesPower,
				"uses_gps":                  input.Cycling.UsesGPS,
				"uses_sports_watch":         input.Cycling.UsesSportsWatch,
				"uses_smart_trainer":        input.Cycling.UsesSmartTrainer,
				"ftp_test_date":             input.Cycling.FTPTestDate,
				"ftp_protocol":              input.Cycling.FTPProtocol,
				"average_power_watts":       input.Cycling.AveragePowerWatts,
				"event_goal":                input.Cycling.EventGoal,
				"event_distance_km":         input.Cycling.EventDistanceKM,
				"event_date":                input.Cycling.EventDate,
			},
			"baseline_eligible": input.BaselineEligible,
			"rotation_index":    input.RotationIndex,
			"observed_training": map[string]any{
				"window_days":              input.Observed.WindowDays,
				"completed_sessions":       input.Observed.CompletedSessions,
				"completed_minutes":        input.Observed.CompletedMinutes,
				"average_rpe":              input.Observed.AverageRPE,
				"average_fatigue":          input.Observed.AverageFatigue,
				"pain_reported":            input.Observed.PainReported,
				"recovery_checkins":        input.Observed.RecoveryCheckins,
				"average_recovery_fatigue": input.Observed.AverageRecoveryFatigue,
				"requires_recovery":        input.Observed.RequiresRecovery(),
			},
		},
		Workouts: workouts,
	}
	return plan, nil
}

func makeWorkout(input Context, slot AvailabilitySlot, kind string, restricted bool, multiplier float64, weekIndex int, date time.Time, eventTaper EventTaperAssessment, postEventRecovery PostEventRecoveryAssessment) Workout {
	if kind == "quality" && input.ExperienceLevel == "beginner" {
		kind = "base"
	}
	baseMinutes := map[string]int{"beginner": 45, "intermediate": 60, "advanced": 75}[input.ExperienceLevel]
	name := "Giro de base"
	targetRPE := 4.0
	mainBlock := "Ritmo confortável e contínuo"
	summary := explanationFor(kind, restricted)
	usesControlledIntervals := false
	usesControlledThreshold := false
	usesRoadModerateIntervals := false
	usesRoadHighIntensityIntervals := false
	usesRoadVO2Intervals := false
	usesShortSelfRegulatedIntervals := false
	usesXCOAerobicIntervals := false
	rotationApplied := false
	activeRecoveryApplied := false
	eventTaperApplied := false
	postEventRecoveryApplied := false
	qualityGoal := qualityGoalFor(input)
	lowCurrentActivity := isLowCurrentActivity(input.Profile.ActivityLevel)
	eventSpecificPhase := eventSpecificPhase(input.Cycling, date)
	observedProtected := input.Observed.RequiresRecovery() && (input.Observed.PainReported || kind == "quality")
	if kind == "base" && weekIndex == 3 {
		name = "Recuperação ativa"
		targetRPE = 3.5
		mainBlock = "Pedale leve e contínuo, sem transformar a sessão em treino de qualidade"
		summary = "A semana de recuperação reduz a carga e usa um giro ativo para manter o movimento sem acrescentar estímulo intenso."
		activeRecoveryApplied = true
	}
	if kind == "long" {
		baseMinutes = map[string]int{"beginner": 75, "intermediate": 90, "advanced": 120}[input.ExperienceLevel]
		name = "Pedal longo"
		targetRPE = 5.0
		mainBlock = "Volume aeróbico estável e sustentável"
	}
	if kind == "quality" {
		name = "Tempo controlado"
		targetRPE = 6.0
		mainBlock = "3 blocos sustentados com recuperação leve"
		preference := preferredQualityPreference(input.Cycling)
		if input.Cycling.Discipline == "road" && input.ExperienceLevel == "advanced" && input.BaselineEligible && (qualityGoal == "performance" || qualityGoal == "event") && input.Cycling.RecentTrainingWeeks >= 8 && input.Cycling.WeeklyRides >= 3 && preference == "vo2max" && slot.AvailableMinutes >= 60 && multiplier >= 0.95 && (!input.Cycling.EventGoal || eventSpecificPhase) {
			name = "Intervalos VO₂max de estrada"
			targetRPE = 8.0
			mainBlock = "4 blocos de 4 min em esforço muito forte-controlado com 4 min leves entre os blocos"
			summary = "A modalidade de estrada, a preferência explícita, o objetivo, a avaliação apta e o histórico mínimo permitem um piloto de VO₂max conservador; a sessão não usa sprint máximo nem meta fixa de potência."
			usesRoadVO2Intervals = true
		} else if (input.Cycling.Discipline == "road" || input.Cycling.Discipline == "indoor") && input.ExperienceLevel == "advanced" && input.BaselineEligible && (qualityGoal == "performance" || qualityGoal == "event") && input.Cycling.RecentTrainingWeeks >= 8 && input.Cycling.WeeklyRides >= 3 && preference == "short_intervals" && slot.AvailableMinutes >= 50 && multiplier >= 0.95 && (!input.Cycling.EventGoal || eventSpecificPhase) {
			name = "Intervalos curtos autorregulados"
			targetRPE = 7.5
			mainBlock = "6 blocos de 1 min em RPE 7–8 com 1 min leve entre os blocos"
			summary = "A modalidade, a preferência explícita, o objetivo, a avaliação apta e o histórico mínimo permitem um piloto curto autorregulado; a sessão não usa sprint máximo, potência fixa ou cadência obrigatória."
			usesShortSelfRegulatedIntervals = true
		} else if (input.Cycling.Discipline == "road" || input.Cycling.Discipline == "indoor") && input.ExperienceLevel == "advanced" && input.BaselineEligible && (qualityGoal == "performance" || qualityGoal == "event") && input.Cycling.RecentTrainingWeeks >= 8 && input.Cycling.WeeklyRides >= 3 && preference == "threshold" && slot.AvailableMinutes >= 60 && multiplier >= 0.95 && (!input.Cycling.EventGoal || eventSpecificPhase) {
			name = "Limiar controlado"
			targetRPE = 7.5
			mainBlock = "3 blocos de 8 min em esforço de limiar controlado com 4 min leves entre os blocos"
			summary = "A modalidade, a preferência explícita, o objetivo, a avaliação apta e o histórico mínimo permitem um piloto de limiar conservador; a sessão usa percepção de esforço e não define potência universal."
			usesControlledThreshold = true
		} else if input.Cycling.Discipline == "road" && input.ExperienceLevel != "beginner" && input.BaselineEligible && (qualityGoal == "performance" || qualityGoal == "event") && slot.AvailableMinutes >= 60 && multiplier >= 0.95 && (preference == "" || preference == "intervals") && (!input.Cycling.EventGoal || eventSpecificPhase) {
			if input.ExperienceLevel == "advanced" && input.Cycling.RecentTrainingWeeks >= 8 && input.Cycling.WeeklyRides >= 3 && input.RotationIndex%2 == 1 && slot.AvailableMinutes >= 75 {
				name = "Intervalos intensos de estrada"
				targetRPE = 8.0
				mainBlock = "5 blocos intensos de 8 min com 4 min leves entre os blocos"
				summary = "A modalidade de estrada, o objetivo, a avaliação apta e a rotação do ciclo permitem um piloto intenso restrito; a sessão não reproduz a carga dos estudos."
				usesRoadHighIntensityIntervals = true
			} else {
				name = "Intervalos moderados de estrada"
				targetRPE = 6.0
				mainBlock = "3 blocos moderados de 10 min com 3 min leves entre os blocos"
				summary = "A modalidade de estrada, o objetivo e a avaliação submáxima apta permitem um piloto intervalado moderado e conservador."
				usesRoadModerateIntervals = true
			}
		} else if input.Cycling.Discipline == "mtb_xco" && input.ExperienceLevel == "advanced" && input.BaselineEligible && (qualityGoal == "performance" || qualityGoal == "event") && slot.AvailableMinutes >= 75 && multiplier >= 0.95 && (preference == "" || preference == "intervals") && (!input.Cycling.EventGoal || eventSpecificPhase) {
			name = "Intervalos aeróbicos XCO"
			targetRPE = 7.0
			mainBlock = "5 blocos aeróbicos de 4 min com 4 min leves entre os blocos"
			summary = "A modalidade XCO explícita, o objetivo compatível e a avaliação submáxima apta liberam um piloto aeróbico conservador; o treino não simula trechos técnicos nem usa sprint máximo."
			usesXCOAerobicIntervals = true
		} else if input.ExperienceLevel == "advanced" && input.Cycling.EventGoal && input.BaselineEligible && eventSpecificPhase {
			name = "Ritmo de prova controlado"
			targetRPE = 6.5
			mainBlock = "3 blocos em ritmo sustentável, com recuperação leve"
			summary = "A prova está próxima o suficiente para orientar um estímulo sustentável, sem simular a prova inteira."
		} else if preference == "cadence" && input.ExperienceLevel != "beginner" {
			name = "Cadência técnica"
			targetRPE = 5.0
			mainBlock = "6 blocos de cadência controlada com recuperação leve"
			summary = "A preferência por cadência orienta uma sessão técnica com esforço controlado."
		} else if preference == "hills" && input.Cycling.Terrain == "hilly" && input.ExperienceLevel != "beginner" {
			name = "Subidas controladas"
			if input.ExperienceLevel == "advanced" {
				targetRPE = 6.5
			}
			mainBlock = "4 blocos sustentados em subida, com recuperação leve"
			summary = "A preferência por subidas e o terreno informado orientam um estímulo controlado."
		} else if preference == "intervals" && input.ExperienceLevel == "advanced" && input.BaselineEligible && (qualityGoal == "performance" || qualityGoal == "event") && slot.AvailableMinutes >= 50 && multiplier >= 0.95 {
			name = "Intervalos controlados"
			targetRPE = 7.0
			mainBlock = "4 blocos de 4 min em esforço forte-controlado, com 3 min leves entre os blocos"
			summary = "A preferência por intervalos foi combinada com uma avaliação submáxima apta e uma semana de construção."
			usesControlledIntervals = true
		} else if preference == "sweet_spot" && input.ExperienceLevel == "advanced" && input.Cycling.UsesPower && input.Cycling.FTP != nil {
			name = "Sweet spot por potência"
			targetRPE = 7.0
			mainBlock = "3 blocos sustentados guiados pelo FTP informado, com recuperação leve"
			summary = "A preferência por sweet spot foi combinada com o medidor de potência e o FTP informado."
		} else if input.ExperienceLevel == "advanced" && input.BaselineEligible && (qualityGoal == "performance" || qualityGoal == "event") && slot.AvailableMinutes >= 50 && multiplier >= 0.95 {
			name = "Intervalos controlados"
			targetRPE = 7.0
			mainBlock = "4 blocos de 4 min em esforço forte-controlado, com 3 min leves entre os blocos"
			summary = "A referência submáxima concluída sem sinais de alerta permite uma progressão intervalada controlada."
			usesControlledIntervals = true
		} else if input.ExperienceLevel == "intermediate" && input.Cycling.BikeType == "indoor" {
			name = "Cadência técnica"
			targetRPE = 5.0
			mainBlock = "6 blocos de cadência controlada com recuperação leve"
			summary = "A sessão usa o ambiente indoor informado para praticar cadência com esforço controlado."
		} else if input.Cycling.Terrain == "hilly" && input.ExperienceLevel != "beginner" {
			name = "Subidas controladas"
			if input.ExperienceLevel == "advanced" {
				targetRPE = 6.5
			}
			mainBlock = "4 blocos sustentados em subida, com recuperação leve"
			summary = "O terreno com subidas informado orienta um estímulo controlado e específico."
		} else if input.ExperienceLevel == "advanced" && input.Cycling.UsesPower && input.Cycling.FTP != nil {
			name = "Sweet spot por potência"
			targetRPE = 7.0
			mainBlock = "3 blocos sustentados guiados pelo FTP informado, com recuperação leve"
			summary = "O medidor de potência e o FTP informados permitem orientar um esforço sustentável."
		} else if input.ExperienceLevel == "advanced" {
			name = "Sweet spot progressivo"
			targetRPE = 7.0
		}
		if input.ExperienceLevel == "advanced" && input.RotationIndex%2 == 1 {
			switch name {
			case "Intervalos controlados":
				name = "Sweet spot progressivo"
				mainBlock = "3 blocos sustentados com recuperação leve"
				summary = "A rotação entre ciclos alterna o estímulo intervalado por um esforço sustentável de qualidade."
				usesControlledIntervals = false
				rotationApplied = true
			case "Subidas controladas", "Sweet spot por potência", "Sweet spot progressivo":
				name = "Tempo controlado"
				targetRPE = 6.0
				mainBlock = "3 blocos sustentados com recuperação leve"
				summary = "A rotação entre ciclos alterna o estímulo específico sem aumentar a carga planejada."
				rotationApplied = true
			}
		}
	}
	if postEventRecoveryAppliesToWorkout(input.Cycling, postEventRecovery, date) {
		postEventRecoveryApplied = true
		name = "Recuperação pós-prova"
		targetRPE = postEventRecoveryTargetRPE
		mainBlock = "Pedale leve e contínuo, sem buscar intensidade após o evento"
		if baseMinutes > postEventRecoveryMaxMinutes {
			baseMinutes = postEventRecoveryMaxMinutes
		}
		summary = "A janela curta após o evento reduz duração e esforço para priorizar recuperação; a evidência disponível não define uma dose universal."
	}
	returningAfterPause := isReturningAfterPause(input.Cycling)
	if returningAfterPause {
		postEventRecoveryApplied = false
		rotationApplied = false
		activeRecoveryApplied = false
		usesControlledIntervals = false
		usesRoadModerateIntervals = false
		usesRoadHighIntensityIntervals = false
		usesRoadVO2Intervals = false
		usesShortSelfRegulatedIntervals = false
		usesXCOAerobicIntervals = false
		name = "Retorno gradual"
		targetRPE = 3.5
		mainBlock = "Pedale leve e contínuo, retomando o ritmo com controle"
		if baseMinutes > 45 {
			baseMinutes = 45
		}
		summary = "O retorno após uma pausa informada recomenda uma retomada gradual; isso não substitui a avaliação de recuperação atual."
	}
	if restricted {
		postEventRecoveryApplied = false
		rotationApplied = false
		activeRecoveryApplied = false
		name = "Giro leve protegido"
		targetRPE = 3.5
		mainBlock = "Esforço leve; interromper diante de dor ou desconforto"
		if baseMinutes > 45 {
			baseMinutes = 45
		}
		multiplier *= 0.8
		summary = explanationFor(kind, true)
	} else if observedProtected {
		postEventRecoveryApplied = false
		rotationApplied = false
		activeRecoveryApplied = false
		name = "Giro leve protegido"
		targetRPE = 3.5
		mainBlock = "Esforço leve; interromper diante de dor ou desconforto"
		if baseMinutes > 45 {
			baseMinutes = 45
		}
		multiplier *= 0.8
		summary = "O histórico recente de esforço, fadiga ou dor recomenda uma sessão leve e protegida neste ciclo."
	}
	duration := int(float64(baseMinutes) * multiplier)
	if duration < 20 {
		duration = 20
	}
	if duration > slot.AvailableMinutes {
		duration = slot.AvailableMinutes
	}
	if !restricted && !observedProtected && eventTaperAppliesToWorkout(input.Cycling, eventTaper, date, weekIndex == 3) {
		eventTaperApplied = true
		duration = int(float64(duration) * eventTaper.VolumeMultiplier)
		if duration < 20 {
			duration = 20
		}
		summary += " O volume foi reduzido para a janela pré-prova, sem aumentar a intensidade nem a frequência planejada."
	}
	protocol := protocolForWorkout(name)

	rules := []string{
		fmt.Sprintf("Agendado em um dia com %d minutos disponíveis.", slot.AvailableMinutes),
		fmt.Sprintf("Carga compatível com experiência %s.", experienceLabel(input.ExperienceLevel)),
		"Progressão de três semanas seguida por uma semana de recuperação.",
	}
	if input.Cycling.WeeklyRides > 0 || input.Cycling.RecentWeeklyDistanceKM > 0 {
		rules = append(rules, "Histórico recente informado usado para contextualizar a sessão.")
	}
	if input.SecondaryGoal != "" && qualityGoal == input.SecondaryGoal && qualityGoal != input.PrimaryGoal {
		rules = append(rules, "Objetivo secundário considerado como desempate para escolher um estímulo de qualidade já elegível.")
	}
	if lowCurrentActivity {
		rules = append(rules, "Rotina atual de baixa atividade: a sessão de qualidade foi preservada até haver mais consistência observada.")
	}
	if slot.PreferredTime != nil || slot.Location != nil {
		rules = append(rules, "Horário e local preferidos foram preservados como contexto da sessão.")
	}
	if input.Observed.HasData() {
		rules = append(rules, fmt.Sprintf("Histórico observado dos últimos %d dias considerado (%d sessões concluídas).", input.Observed.WindowDays, input.Observed.CompletedSessions))
	}
	if hasLowObservedAdherence(input.TrainingHistory) {
		rules = append(rules, "Baixa aderência observada: a sessão de qualidade não foi incluída para reduzir complexidade e favorecer a retomada da consistência.")
	}
	if preference := preferredQualityPreference(input.Cycling); kind == "quality" && preference != "" {
		rules = append(rules, fmt.Sprintf("Preferência por %s considerada dentro dos limites de segurança.", sessionPreferenceLabel(preference)))
	}
	if usesControlledIntervals {
		rules = append(rules, "Intervalos liberados pela avaliação submáxima apta, apenas nas semanas de construção.")
	}
	if usesControlledThreshold {
		rules = append(rules, "Piloto de limiar liberado por preferência explícita, modalidade, objetivo, avaliação apta e histórico mínimo; esforço guiado por RPE, sem potência universal.")
	}
	if usesRoadModerateIntervals {
		rules = append(rules, "Piloto de estrada moderado liberado por modalidade explícita, objetivo compatível, avaliação apta e disponibilidade suficiente.")
	}
	if usesRoadHighIntensityIntervals {
		rules = append(rules, "Piloto intenso de estrada liberado apenas para atleta avançado elegível, em ciclo alternado e semana de construção; sem reprodução da carga estudada.")
	}
	if usesRoadVO2Intervals {
		rules = append(rules, "Piloto de VO₂max de estrada liberado por preferência explícita, modalidade, objetivo, avaliação apta e histórico mínimo; sem sprint máximo ou meta fixa de potência.")
	}
	if usesShortSelfRegulatedIntervals {
		rules = append(rules, "Piloto de intervalos curtos liberado por preferência explícita, modalidade, objetivo, avaliação apta e histórico mínimo; esforço autorregulado, sem sprint máximo ou meta fixa de potência.")
	}
	if usesXCOAerobicIntervals {
		rules = append(rules, "Piloto aeróbico XCO liberado por modalidade explícita, objetivo compatível, avaliação apta e disponibilidade suficiente; sem sprint máximo ou simulação técnica.")
	}
	if rotationApplied {
		rules = append(rules, "Sessão alternada pela rotação explicável do ciclo, sem aumentar a carga planejada.")
	}
	if activeRecoveryApplied {
		rules = append(rules, "Variação de recuperação ativa aplicada na quarta semana, sem aumentar a carga planejada.")
	}
	if returningAfterPause && name == "Retorno gradual" {
		rules = append(rules, "O retorno após uma pausa informada mantém a retomada leve, limitada a 45 minutos e RPE 3,5, sem sessão de qualidade.")
	}
	if eventTaperApplied {
		rules = append(rules, "Taper pré-prova aplicado nesta sessão: volume reduzido de forma conservadora, mantendo a frequência planejada.")
	}
	if postEventRecoveryApplied {
		rules = append(rules, "Recuperação pós-prova aplicada por até sete dias após o evento: duração limitada a 45 minutos e RPE 3,5, sem estímulo de qualidade.")
	}
	if restricted {
		rules = append(rules, "Intensidade limitada por uma condição de segurança ativa.")
	}
	if hasMedicalRestriction(input.Limitations) {
		rules = append(rules, "Restrição médica informada: a carga permanece protegida e não substitui orientação profissional.")
	}
	if observedProtected {
		rules = append(rules, "Sessão protegida por sinais recentes de recuperação insuficiente ou dor relatada.")
	}
	evidenceKeys := append([]string(nil), protocol.EvidenceKeys...)
	evidenceScope := protocol.EvidenceScope
	if eventTaperApplied {
		evidenceKeys = append(append([]string(nil), eventTaper.EvidenceKeys...), evidenceKeys...)
		evidenceScope += " O taper pré-prova usa evidência de redução de volume em ciclistas/endurance, com transferência limitada a atletas elegíveis; não é dose universal."
	}
	structure := buildStructure(duration, targetRPE, name, mainBlock)
	if slot.PreferredTime != nil {
		structure["preferred_time"] = *slot.PreferredTime
	}
	if slot.Location != nil {
		structure["planned_location"] = *slot.Location
	}
	decisionAudit := buildWorkoutDecisionAudit(input, kind, rules, restricted, observedProtected, returningAfterPause, weekIndex == 3, eventTaperApplied, postEventRecoveryApplied)
	return Workout{
		ScheduledOn:     date.Format("2006-01-02"),
		Name:            name,
		Objective:       objectiveFor(input.PrimaryGoal),
		DurationMinutes: duration,
		TargetRPE:       targetRPE,
		Structure:       structure,
		Explanation:     map[string]any{"summary": summary, "rules": rules, "decision_audit": decisionAudit, "protocol_key": protocol.Key, "protocol_metadata": metadataForProtocol(protocol.Key), "evidence_keys": evidenceKeys, "evidence_scope": evidenceScope, "event_taper_applied": eventTaperApplied},
		Status:          "planned",
	}
}
