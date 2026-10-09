package activityimport

import (
	"testing"
	"time"
)

func intp(v int) *int { return &v }

func activityOn(day int, hr, power *int) Activity {
	return Activity{Parsed: Parsed{StartedAt: time.Date(2026, 9, day, 8, 0, 0, 0, time.UTC), MaxHeartRate: hr, Best20MinPowerW: power}}
}

func TestSuggestionsNeedEnoughActivities(t *testing.T) {
	got := suggestionsFrom([]Activity{activityOn(1, intp(180), intp(200)), activityOn(2, intp(185), intp(210))})
	if got.MaxHeartRate != nil || got.FTP != nil {
		t.Fatalf("com 2 atividades não deve sugerir, veio %+v", got)
	}
}

func TestSuggestionsUseHighestHeartRateAndNinetyFivePercentOfBest20Minutes(t *testing.T) {
	got := suggestionsFrom([]Activity{
		activityOn(1, intp(180), intp(200)),
		activityOn(5, intp(191), intp(240)),
		activityOn(9, intp(176), intp(220)),
	})
	if got.MaxHeartRate == nil || got.MaxHeartRate.Value != 191 || got.MaxHeartRate.ObservedOn != "2026-09-05" || got.MaxHeartRate.Activities != 3 {
		t.Fatalf("FC máxima inesperada: %+v", got.MaxHeartRate)
	}
	if got.FTP == nil || got.FTP.Value != 228 || got.FTP.ObservedOn != "2026-09-05" {
		t.Fatalf("FTP inesperado: %+v", got.FTP)
	}
}

func TestSuggestionsIgnoreImplausibleHeartRate(t *testing.T) {
	got := suggestionsFrom([]Activity{
		activityOn(1, intp(250), nil), activityOn(2, intp(60), nil), activityOn(3, intp(182), nil),
		activityOn(4, intp(185), nil),
	})
	if got.MaxHeartRate != nil {
		t.Fatalf("só 2 leituras plausíveis, não deveria sugerir: %+v", got.MaxHeartRate)
	}
}

func TestBestAveragePowerFindsStrongestWindow(t *testing.T) {
	start := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	var samples []powerSample
	for second := 0; second < 3000; second++ {
		watts := 150
		if second >= 1000 && second < 2200 {
			watts = 250
		}
		samples = append(samples, powerSample{at: start.Add(time.Duration(second) * time.Second), watts: watts})
	}
	best := bestAveragePower(samples, 1200)
	if best == nil || *best != 250 {
		t.Fatalf("melhor janela = %v, quer 250", best)
	}
	if bestAveragePower(samples[:600], 1200) != nil {
		t.Fatal("arquivo mais curto que a janela não deve gerar valor")
	}
}

func TestBestAveragePowerTreatsLongGapsAsZero(t *testing.T) {
	start := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	samples := []powerSample{{at: start, watts: 300}, {at: start.Add(1500 * time.Second), watts: 300}}
	best := bestAveragePower(samples, 1200)
	if best != nil && *best > 20 {
		t.Fatalf("pausa longa não pode virar janela forte, veio %d", *best)
	}
}
