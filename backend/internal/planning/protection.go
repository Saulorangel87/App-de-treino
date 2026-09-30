package planning

import (
	"encoding/json"
	"sort"
	"time"
)

// ProtectionLevel is how much the recent history should restrain upcoming
// sessions. It replaces the all-or-nothing Observed.RequiresRecovery() once it
// is wired into prescription; until then it is only assessed and tested.
type ProtectionLevel string

const (
	ProtectionNone     ProtectionLevel = "none"
	ProtectionLight    ProtectionLevel = "light"
	ProtectionModerate ProtectionLevel = "moderate"
	ProtectionStrong   ProtectionLevel = "strong"
)

const (
	protectionPainStrongDays    = 3  // dor isolada: forte nos dias 0 a 2
	protectionPainModerateDays  = 7  // dor isolada: moderada até o 7º dia
	protectionRecurrenceWindow  = 14 // janela para contar dor recorrente
	protectionRecurrenceStrong  = 7  // dor recorrente: forte por 7 dias
	protectionFatigueWindowDays = 7
	protectionGoodCheckinAfter  = 3 // check-in bom só rebaixa a partir do 3º dia após a dor
	protectionGoodFatigueMax    = 2
	protectionHighFatigueMin    = 4
)

// RecentSignal is one dated record from a completed session's feedback, from a
// recovery check-in or from the athlete's "I am recovered" declaration
// ("self_report", stored with fatigue 1). Fatigue is 0 when the record has no
// fatigue value.
type RecentSignal struct {
	Date         time.Time
	Source       string // "session", "checkin" or "self_report"
	PainReported bool
	Fatigue      int
}

type ProtectionAssessment struct {
	Level ProtectionLevel `json:"level"`
	// ExpiresOn is the first date on which the signals stop applying, assuming no
	// new records. Empty when the level is none.
	ExpiresOn           string            `json:"expires_on,omitempty"`
	SuggestProfessional bool              `json:"suggest_professional"`
	Reasons             []ReadinessReason `json:"reasons"`
}

// ProtectionEffect is what a level asks of the prescription.
type ProtectionEffect struct {
	DurationFactor float64
	// QualityRPEDelta reduces the target RPE of quality sessions (light only).
	QualityRPEDelta float64
	// ReplaceQuality swaps quality sessions for easy aerobic riding.
	ReplaceQuality bool
	// ProtectedSession turns every affected session into the protected easy ride.
	ProtectedSession bool
}

func (level ProtectionLevel) Effect() ProtectionEffect {
	switch level {
	case ProtectionLight:
		return ProtectionEffect{DurationFactor: 0.9, QualityRPEDelta: -1}
	case ProtectionModerate:
		return ProtectionEffect{DurationFactor: 0.9, ReplaceQuality: true}
	case ProtectionStrong:
		return ProtectionEffect{DurationFactor: 0.8, ProtectedSession: true}
	default:
		return ProtectionEffect{DurationFactor: 1}
	}
}

var protectionOrder = []ProtectionLevel{ProtectionNone, ProtectionLight, ProtectionModerate, ProtectionStrong}

var protectionRank = map[ProtectionLevel]int{ProtectionNone: 0, ProtectionLight: 1, ProtectionModerate: 2, ProtectionStrong: 3}

func lowerProtection(level ProtectionLevel) ProtectionLevel {
	return protectionOrder[max(protectionRank[level]-1, 0)]
}

func maxProtection(a, b ProtectionLevel) ProtectionLevel {
	if protectionRank[b] > protectionRank[a] {
		return b
	}
	return a
}

// dayIndex counts whole UTC days so that "age in days" does not depend on the
// time of day a record was stored. The athlete's timezone is not evaluated.
func dayIndex(t time.Time) int {
	utc := t.UTC()
	return int(time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC).Unix() / 86400)
}

func protectionDate(day int) string {
	return time.Unix(int64(day)*86400, 0).UTC().Format("2006-01-02")
}

// assessProtection is a pure function of dated signals and today's date. Signals
// dated in the future are ignored.
func assessProtection(signals []RecentSignal, now time.Time) ProtectionAssessment {
	today := dayIndex(now)
	result := ProtectionAssessment{Level: ProtectionNone, Reasons: []ReadinessReason{}}

	type aged struct {
		RecentSignal
		age int
	}
	var records []aged
	for _, signal := range signals {
		age := today - dayIndex(signal.Date)
		if age < 0 {
			continue
		}
		records = append(records, aged{signal, age})
	}
	// Newest first; ties keep the input order.
	sort.SliceStable(records, func(i, j int) bool { return records[i].age < records[j].age })

	addReason := func(code, message string) {
		result.Reasons = append(result.Reasons, ReadinessReason{Code: code, Message: message})
	}
	expire := 0 // day index of the latest expiry among the signals that apply
	setExpiry := func(day int) { expire = max(expire, day) }

	// Pain.
	var pains []aged
	for _, record := range records {
		if record.PainReported {
			pains = append(pains, record)
		}
	}
	painLevel := ProtectionNone
	if len(pains) > 0 {
		latest := pains[0]
		recent := 0
		for _, pain := range pains {
			if pain.age <= protectionRecurrenceWindow {
				recent++
			}
		}
		switch {
		case recent >= 2 && latest.age < protectionRecurrenceStrong:
			painLevel = ProtectionStrong
			result.SuggestProfessional = true
			setExpiry(today - latest.age + protectionRecurrenceStrong)
			addReason("recurrent_pain", "Dor relatada mais de uma vez em 14 dias; proteção forte por 7 dias e recomendação de avaliação profissional se a dor persistir.")
		case latest.age < protectionPainStrongDays:
			painLevel = ProtectionStrong
			setExpiry(today - latest.age + protectionPainStrongDays)
			addReason("recent_pain", "Dor relatada nos últimos dias; proteção forte no início e reduzida nos dias seguintes.")
		case latest.age <= protectionPainModerateDays:
			painLevel = ProtectionModerate
			setExpiry(today - latest.age + protectionPainModerateDays + 1)
			addReason("pain_within_week", "Dor relatada há menos de uma semana; a proteção diminui conforme os dias passam.")
		}
		// Recurrent pain is never lowered by later good records.
		if painLevel != ProtectionStrong || !result.SuggestProfessional {
			goodAfter := false
			for _, record := range records {
				if (record.Source == "checkin" || record.Source == "self_report") && !record.PainReported && record.Fatigue >= 1 && record.Fatigue <= protectionGoodFatigueMax &&
					record.age <= latest.age-protectionGoodCheckinAfter {
					goodAfter = true
					break
				}
			}
			if goodAfter && painLevel != ProtectionNone {
				painLevel = maxProtection(ProtectionLight, lowerProtection(painLevel))
				addReason("recovered_checkin_after_pain", "Check-in posterior à dor sem dor e com fadiga baixa reduziu a proteção em um nível.")
			}
		}
	}

	// Fatigue (sessions and check-ins), newest record decides whether it recovered.
	var fatigued []aged
	for _, record := range records {
		if record.Fatigue > 0 {
			fatigued = append(fatigued, record)
		}
	}
	fatigueLevel := ProtectionNone
	if len(fatigued) > 0 {
		var highs []aged
		for _, record := range fatigued {
			if record.Fatigue >= protectionHighFatigueMin && record.age <= protectionFatigueWindowDays {
				highs = append(highs, record)
			}
		}
		persistent := false
		var sessionsRun []aged
		for _, record := range fatigued {
			if record.Source == "session" && record.age <= protectionRecurrenceWindow {
				sessionsRun = append(sessionsRun, record)
			}
		}
		if len(sessionsRun) >= 3 {
			persistent = true
			for _, record := range sessionsRun[:3] {
				if record.Fatigue < protectionHighFatigueMin {
					persistent = false
				}
			}
		}
		if len(highs) > 0 {
			latestHigh := highs[0]
			recovered := false
			for _, record := range fatigued {
				if record.age < latestHigh.age && record.Fatigue <= protectionGoodFatigueMax {
					recovered = true
					break
				}
			}
			switch {
			case recovered:
				// A newer, clearly good record outweighs earlier high fatigue.
			case persistent || len(highs) >= 2:
				fatigueLevel = ProtectionModerate
				setExpiry(today - latestHigh.age + protectionFatigueWindowDays + 1)
				addReason("repeated_high_fatigue", "Fadiga alta repetida na última semana; os treinos de qualidade são trocados por pedal aeróbico leve.")
			default:
				fatigueLevel = ProtectionLight
				setExpiry(today - latestHigh.age + protectionFatigueWindowDays + 1)
				addReason("high_fatigue", "Fadiga alta registrada na última semana; carga levemente reduzida.")
			}
		}
	}

	result.Level = maxProtection(painLevel, fatigueLevel)
	if result.Level != ProtectionNone && expire > 0 {
		result.ExpiresOn = protectionDate(expire)
	}
	if result.Level == ProtectionNone {
		result.Reasons = []ReadinessReason{}
	}
	return result
}

// PrescriptionInputs are the per-workout parameters makeWorkout needs to build
// the same session again. They are stored in the workout's explanation so that
// upcoming sessions can be re-evaluated (and restored) without regenerating the
// whole plan. Workouts generated before this field existed cannot be rebuilt.
type PrescriptionInputs struct {
	Kind       string  `json:"kind"`
	WeekIndex  int     `json:"week_index"`
	Multiplier float64 `json:"multiplier"`
	Weekday    int     `json:"weekday"`
}

func prescriptionInputsFrom(explanation map[string]any) (PrescriptionInputs, bool) {
	raw, ok := explanation["prescription_inputs"]
	if !ok {
		return PrescriptionInputs{}, false
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return PrescriptionInputs{}, false
	}
	var inputs PrescriptionInputs
	if err := json.Unmarshal(encoded, &inputs); err != nil || inputs.Kind == "" || inputs.Multiplier <= 0 {
		return PrescriptionInputs{}, false
	}
	return inputs, true
}

// requiresRecovery keeps the legacy aggregate rule until a protection
// assessment is attached to the context.
func (input Context) requiresRecovery() bool {
	if input.Protection != nil {
		return input.Protection.Level != ProtectionNone
	}
	return input.Observed.RequiresRecovery()
}

// sessionProtection reports whether a session of the given kind becomes the
// protected easy ride, and which graduated level applies. A nil Protection is the
// legacy rule: pain protects every session, high fatigue only quality sessions.
func (input Context) sessionProtection(kind string) (protected bool, level ProtectionLevel) {
	if input.Protection == nil {
		return input.Observed.RequiresRecovery() && (input.Observed.PainReported || kind == "quality"), ""
	}
	level = input.Protection.Level
	switch level {
	case ProtectionStrong:
		return true, level
	case ProtectionModerate:
		return kind == "quality", level
	}
	return false, level
}

func protectionRule(level ProtectionLevel) string {
	switch level {
	case ProtectionLight:
		return "Proteção leve por sinal recente de recuperação: duração reduzida em 10% e esforço dos treinos de qualidade reduzido em 1 ponto."
	case ProtectionModerate:
		return "Proteção moderada por sinais recentes de recuperação: duração reduzida em 10%; treinos de qualidade são trocados por giro leve protegido."
	}
	return ""
}

func (assessment ProtectionAssessment) explanation() map[string]any {
	reasons := make([]string, 0, len(assessment.Reasons))
	for _, reason := range assessment.Reasons {
		reasons = append(reasons, reason.Message)
	}
	result := map[string]any{
		"level":                assessment.Level,
		"reasons":              reasons,
		"suggest_professional": assessment.SuggestProfessional,
	}
	if assessment.ExpiresOn != "" {
		result["expires_on"] = assessment.ExpiresOn
	}
	return result
}

// restrictionState reports the profile-level restrictions that always win over
// history-based protection.
func restrictionState(input Context) (restricted, medical bool) {
	restricted = len(input.Limitations) > 0
	for _, item := range input.Limitations {
		if item.ProfessionalClearanceRecommended {
			restricted = true
		}
		if item.MedicalRestriction {
			medical = true
		}
	}
	return restricted, medical
}

// ReprescribeWorkout rebuilds one planned session with the current context. It
// returns false when the session cannot be rebuilt (no stored inputs, or the
// weekday is no longer an available slot), in which case it must be left as is.
func ReprescribeWorkout(input Context, scheduledOn string, explanation map[string]any, now time.Time) (Workout, bool) {
	inputs, ok := prescriptionInputsFrom(explanation)
	if !ok {
		return Workout{}, false
	}
	date, err := time.ParseInLocation("2006-01-02", scheduledOn, now.Location())
	if err != nil {
		return Workout{}, false
	}
	var slot *AvailabilitySlot
	for index := range input.Availability {
		if input.Availability[index].Weekday == inputs.Weekday {
			slot = &input.Availability[index]
			break
		}
	}
	if slot == nil {
		return Workout{}, false
	}
	restricted, _ := restrictionState(input)
	return makeWorkout(input, *slot, inputs.Kind, restricted, inputs.Multiplier, inputs.WeekIndex, date,
		assessEventTaper(input, now, restricted), assessPostEventRecovery(input, now)), true
}
