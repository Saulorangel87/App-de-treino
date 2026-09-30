package planning

import (
	"testing"
	"time"
)

// Caracterização do comportamento ATUAL da proteção por histórico observado.
// O motor só enxerga agregados de 28 dias (PainReported, médias de fadiga), então
// cenários que diferem no tempo (dor de ontem x dor de 27 dias atrás, melhora
// depois do sinal) chegam ao motor com exatamente o mesmo Observed. Os números
// abaixo devem mudar junto com a regra quando a proteção ganhar validade e
// níveis; o comportamento desejado está em docs/motor-protecao-cenarios.md.

type protectionOutcome struct {
	protected, total, quality int
	maxRPE                    float64
	maxMinutes                int
}

func protectionOutcomeFor(t *testing.T, input Context) protectionOutcome {
	t.Helper()
	plan, err := buildPlan(input, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("buildPlan: %v", err)
	}
	var out protectionOutcome
	for _, workout := range plan.Workouts {
		out.total++
		if workout.Name == "Giro leve protegido" {
			out.protected++
		}
		if workout.TargetRPE >= 6 {
			out.quality++
		}
		out.maxRPE = max(out.maxRPE, workout.TargetRPE)
		out.maxMinutes = max(out.maxMinutes, workout.DurationMinutes)
	}
	return out
}

func TestCurrentProtectionScenarios(t *testing.T) {
	scenarios := []struct {
		name                       string
		change                     func(*Context)
		wantProtected, wantQuality int
		wantMaxRPE                 float64
		wantMaxMinutes             int
	}{
		{"sem sinais", func(*Context) {}, 0, 2, 6.0, 120},
		// Dor de ontem e dor de 27 dias atrás são indistinguíveis: o motor recebe
		// apenas PainReported=true e protege o ciclo inteiro.
		{"dor relatada em sessão, de qualquer idade dentro de 28 dias", func(c *Context) { c.Observed.PainReported = true }, 11, 0, 3.5, 36},
		{"fadiga média de sessões 3,9, abaixo do limiar", func(c *Context) { c.Observed.AverageFatigue = 3.9 }, 0, 2, 6.0, 120},
		// Fadiga alta troca só os treinos de qualidade; os demais permanecem.
		{"fadiga média de sessões 4,0", func(c *Context) { c.Observed.AverageFatigue = 4 }, 2, 0, 5.0, 120},
		{"fadiga média de check-ins 4,0", func(c *Context) { c.Observed.AverageRecoveryFatigue = 4 }, 2, 0, 5.0, 120},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			input := readinessContext()
			scenario.change(&input)
			got := protectionOutcomeFor(t, input)
			want := protectionOutcome{
				protected: scenario.wantProtected, total: 11, quality: scenario.wantQuality,
				maxRPE: scenario.wantMaxRPE, maxMinutes: scenario.wantMaxMinutes,
			}
			if got != want {
				t.Fatalf("got %+v, want %+v", got, want)
			}
		})
	}
}
