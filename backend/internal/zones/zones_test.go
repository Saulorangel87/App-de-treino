package zones

import (
	"os"
	"regexp"
	"strconv"
	"testing"
)

func TestForRPEUsesTheZoneBoundaries(t *testing.T) {
	cases := map[float64]int{
		1: 1, 3: 1, 3.5: 1, 3.6: 2, 4: 2, 5: 2, 5.5: 3, 6: 3, 6.5: 3, 7: 4, 7.5: 4, 8: 5, 9.5: 5, 10: 5, 12: 5,
	}
	for rpe, want := range cases {
		if got := ForRPE(rpe); got != want {
			t.Errorf("ForRPE(%v) = %d, want %d", rpe, got, want)
		}
	}
}

func TestEveryRPEThePlannerPrescribesLandsInTheIntendedZone(t *testing.T) {
	// Target RPEs used by the planner, by kind of session.
	want := map[string]struct {
		rpe  float64
		zone int
	}{
		"recuperação e giro leve protegido": {3.5, 1},
		"giro de base":                      {4, 2},
		"endurance e pedal longo":           {5, 2},
		"tempo e intervalos moderados":      {6, 3},
		"subidas controladas":               {6.5, 3},
		"sweet spot":                        {7, 4},
		"intervalos de limiar":              {7.5, 4},
		"intervalos intensos":               {8, 5},
	}
	for name, c := range want {
		if got := ForRPE(c.rpe); got != c.zone {
			t.Errorf("%s (RPE %v) is zone %d, want %d", name, c.rpe, got, c.zone)
		}
	}
}

func TestReportedZoneRoundTrips(t *testing.T) {
	for _, zone := range All {
		rpe, ok := RPEForZone(zone.Number)
		if !ok || ForRPE(rpe) != zone.Number {
			t.Errorf("zone %d: stored RPE %v maps back to zone %d", zone.Number, rpe, ForRPE(rpe))
		}
	}
	for _, invalid := range []int{0, -1, 6, 99} {
		if _, ok := RPEForZone(invalid); ok {
			t.Errorf("zone %d must not exist", invalid)
		}
	}
}

func TestLabel(t *testing.T) {
	if got := Label(4.5); got != "Z2 · Resistência" {
		t.Errorf("Label(4.5) = %q", got)
	}
}

func TestTheFrontendTableMatchesTheBackend(t *testing.T) {
	source, err := os.ReadFile("../../../frontend/lib/zones.ts")
	if err != nil {
		t.Skipf("frontend sources are not available: %v", err)
	}
	pattern := regexp.MustCompile(`number: (\d), name: '([^']+)', maxRpe: ([0-9.]+), feedbackRpe: ([0-9.]+)`)
	matches := pattern.FindAllStringSubmatch(string(source), -1)
	if len(matches) != len(All) {
		t.Fatalf("frontend/lib/zones.ts declares %d zones, the backend %d", len(matches), len(All))
	}
	for index, match := range matches {
		number, _ := strconv.Atoi(match[1])
		maxRPE, _ := strconv.ParseFloat(match[3], 64)
		feedbackRPE, _ := strconv.ParseFloat(match[4], 64)
		zone := All[index]
		if number != zone.Number || match[2] != zone.Name || maxRPE != zone.MaxRPE || feedbackRPE != zone.FeedbackRPE {
			t.Errorf("zone %d differs between frontend (%v) and backend (%+v)", zone.Number, match[1:], zone)
		}
	}
}
