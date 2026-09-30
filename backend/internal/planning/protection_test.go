package planning

import (
	"testing"
	"time"
)

var protectionNow = time.Date(2026, time.September, 30, 10, 0, 0, 0, time.UTC)

// daysAgo builds a signal dated n days before protectionNow.
func daysAgo(n int, source string, pain bool, fatigue int) RecentSignal {
	return RecentSignal{Date: protectionNow.AddDate(0, 0, -n), Source: source, PainReported: pain, Fatigue: fatigue}
}

// Cada linha corresponde a um cenário de docs/motor-protecao-cenarios.md.
func TestAssessProtectionScenarios(t *testing.T) {
	tests := []struct {
		name         string
		signals      []RecentSignal
		want         ProtectionLevel
		professional bool
		expiresOn    string
	}{
		{"1 sem sinais", nil, ProtectionNone, false, ""},
		{"1 registros normais", []RecentSignal{daysAgo(1, "session", false, 2), daysAgo(2, "checkin", false, 2)}, ProtectionNone, false, ""},
		{"2 dor ontem", []RecentSignal{daysAgo(1, "session", true, 3)}, ProtectionStrong, false, "2026-10-02"},
		{"2 dor hoje", []RecentSignal{daysAgo(0, "session", true, 3)}, ProtectionStrong, false, "2026-10-03"},
		{"2 dor há 3 dias vira moderada", []RecentSignal{daysAgo(3, "session", true, 3)}, ProtectionModerate, false, "2026-10-05"},
		{"2 dor há 7 dias ainda moderada", []RecentSignal{daysAgo(7, "session", true, 3)}, ProtectionModerate, false, "2026-10-01"},
		{"3 dor há 8 dias", []RecentSignal{daysAgo(8, "session", true, 3)}, ProtectionNone, false, ""},
		{"3 dor há 20 dias", []RecentSignal{daysAgo(20, "session", true, 3)}, ProtectionNone, false, ""},
		{"4 dor recorrente", []RecentSignal{daysAgo(1, "session", true, 3), daysAgo(9, "session", true, 3)}, ProtectionStrong, true, "2026-10-06"},
		{"4 dor recorrente continua forte com check-in bom", []RecentSignal{daysAgo(3, "session", true, 3), daysAgo(9, "session", true, 3), daysAgo(0, "checkin", false, 1)}, ProtectionStrong, true, "2026-10-04"},
		{"4 duas dores mas a mais recente tem 10 dias", []RecentSignal{daysAgo(10, "session", true, 3), daysAgo(12, "session", true, 3)}, ProtectionNone, false, ""},
		{"5 dor há 5 dias com check-in bom depois", []RecentSignal{daysAgo(5, "session", true, 3), daysAgo(1, "checkin", false, 2), daysAgo(0, "checkin", false, 1)}, ProtectionLight, false, "2026-10-03"},
		{"5 check-in bom cedo demais não rebaixa", []RecentSignal{daysAgo(4, "session", true, 3), daysAgo(3, "checkin", false, 1)}, ProtectionModerate, false, "2026-10-04"},
		{"2 dor ontem com check-in bom hoje continua forte", []RecentSignal{daysAgo(1, "session", true, 3), daysAgo(0, "checkin", false, 1)}, ProtectionStrong, false, "2026-10-02"},
		{"6 um dia de fadiga 5 já seguido de registros bons", []RecentSignal{daysAgo(4, "session", false, 5), daysAgo(3, "session", false, 2), daysAgo(1, "session", false, 2)}, ProtectionNone, false, ""},
		{"6 fadiga 5 como registro mais recente", []RecentSignal{daysAgo(1, "session", false, 5), daysAgo(3, "session", false, 2), daysAgo(5, "session", false, 2)}, ProtectionLight, false, "2026-10-07"},
		{"7 fadiga 5 e depois 3", []RecentSignal{daysAgo(2, "session", false, 5), daysAgo(1, "session", false, 3)}, ProtectionLight, false, "2026-10-06"},
		{"8 fadiga de check-in alta e melhora nos dois mais recentes", []RecentSignal{daysAgo(4, "checkin", false, 4), daysAgo(1, "checkin", false, 2), daysAgo(0, "checkin", false, 1)}, ProtectionNone, false, ""},
		{"9 três sessões seguidas com fadiga alta", []RecentSignal{daysAgo(1, "session", false, 4), daysAgo(3, "session", false, 5), daysAgo(5, "session", false, 4)}, ProtectionModerate, false, "2026-10-07"},
		{"9 duas altas na semana", []RecentSignal{daysAgo(1, "checkin", false, 4), daysAgo(4, "session", false, 5)}, ProtectionModerate, false, "2026-10-07"},
		{"fadiga alta há mais de 7 dias não conta", []RecentSignal{daysAgo(8, "session", false, 5)}, ProtectionNone, false, ""},
		{"registro no futuro é ignorado", []RecentSignal{{Date: protectionNow.AddDate(0, 0, 2), Source: "session", PainReported: true, Fatigue: 5}}, ProtectionNone, false, ""},
		{"dor forte vence fadiga leve; o prazo final é o da fadiga", []RecentSignal{daysAgo(1, "session", true, 4)}, ProtectionStrong, false, "2026-10-07"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := assessProtection(test.signals, protectionNow)
			if got.Level != test.want || got.SuggestProfessional != test.professional || got.ExpiresOn != test.expiresOn {
				t.Fatalf("got level=%s professional=%v expires=%q reasons=%v; want level=%s professional=%v expires=%q",
					got.Level, got.SuggestProfessional, got.ExpiresOn, got.Reasons, test.want, test.professional, test.expiresOn)
			}
			if (got.Level == ProtectionNone) != (len(got.Reasons) == 0) {
				t.Fatalf("reasons must be present exactly when protection applies: %+v", got)
			}
		})
	}
}

// Comparação com o motor atual: onde o agregado de 28 dias protege o ciclo
// inteiro, o novo sinal diferencia pelo tempo e pelas melhoras posteriores.
func TestAssessProtectionDiffersFromCurrentRule(t *testing.T) {
	old := func(pain bool, sessionFatigue, checkinFatigue float64) bool {
		return ObservedTrainingSummary{PainReported: pain, AverageFatigue: sessionFatigue, AverageRecoveryFatigue: checkinFatigue}.RequiresRecovery()
	}
	cases := []struct {
		name            string
		signals         []RecentSignal
		pain            bool
		sessionFatigue  float64
		checkinFatigue  float64
		wantOld, wantIn ProtectionLevel
	}{
		{"dor há 20 dias", []RecentSignal{daysAgo(20, "session", true, 3)}, true, 3, 0, ProtectionStrong, ProtectionNone},
		{"fadiga de check-in com melhora", []RecentSignal{daysAgo(4, "checkin", false, 4), daysAgo(0, "checkin", false, 1)}, false, 0, 2.5, ProtectionNone, ProtectionNone},
		{"dor há 5 dias com melhora", []RecentSignal{daysAgo(5, "session", true, 3), daysAgo(1, "checkin", false, 2)}, true, 3, 2, ProtectionStrong, ProtectionLight},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			oldProtected := old(test.pain, test.sessionFatigue, test.checkinFatigue)
			if oldProtected != (test.wantOld != ProtectionNone) {
				t.Fatalf("current rule protected=%v", oldProtected)
			}
			if got := assessProtection(test.signals, protectionNow).Level; got != test.wantIn {
				t.Fatalf("new level=%s, want %s", got, test.wantIn)
			}
		})
	}
}

func TestProtectionEffectsAreOrdered(t *testing.T) {
	previous := 1.1
	for _, level := range protectionOrder {
		effect := level.Effect()
		if effect.DurationFactor > previous || effect.DurationFactor <= 0 {
			t.Fatalf("%s duration factor %.2f must not exceed the previous level's %.2f", level, effect.DurationFactor, previous)
		}
		previous = effect.DurationFactor
	}
	if none := ProtectionNone.Effect(); none.DurationFactor != 1 || none.ReplaceQuality || none.ProtectedSession || none.QualityRPEDelta != 0 {
		t.Fatalf("none must not change anything: %+v", none)
	}
	if !ProtectionStrong.Effect().ProtectedSession || !ProtectionModerate.Effect().ReplaceQuality || ProtectionModerate.Effect().ProtectedSession {
		t.Fatal("effects do not match the documented levels")
	}
}
