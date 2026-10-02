// Package zones holds the five effort zones the athlete sees (Z1 to Z5) and how
// they relate to the RPE scale the engine still prescribes and adapts with. The
// engine, the stored workouts and the scientific evidence keep speaking RPE; the
// zones are the language of the app. The limits between zones are a product
// choice, not a finding from a study, and must match frontend/lib/zones.ts (a
// test enforces it).
package zones

import "fmt"

// Zone is one effort zone.
type Zone struct {
	Number int
	Name   string
	// MaxRPE is the highest prescribed RPE that still belongs to the zone; the
	// last zone has no upper limit.
	MaxRPE float64
	// FeedbackRPE is the RPE stored when the athlete reports having ridden in the
	// zone. It lies inside the zone, so ForRPE(FeedbackRPE) is the same zone.
	FeedbackRPE float64
}

// All lists the zones from the easiest to the hardest.
var All = []Zone{
	{Number: 1, Name: "Recuperação", MaxRPE: 3.5, FeedbackRPE: 3},
	{Number: 2, Name: "Resistência", MaxRPE: 5, FeedbackRPE: 4.5},
	{Number: 3, Name: "Ritmo", MaxRPE: 6.5, FeedbackRPE: 6},
	{Number: 4, Name: "Limiar", MaxRPE: 7.5, FeedbackRPE: 7},
	{Number: 5, Name: "Intenso", MaxRPE: 10, FeedbackRPE: 8.5},
}

// ForRPE returns the zone (1 to 5) of a prescribed or reported RPE.
func ForRPE(rpe float64) int {
	for _, zone := range All {
		if rpe <= zone.MaxRPE {
			return zone.Number
		}
	}
	return All[len(All)-1].Number
}

// RPEForZone returns the RPE stored for a reported zone, and false when the zone
// does not exist.
func RPEForZone(number int) (float64, bool) {
	for _, zone := range All {
		if zone.Number == number {
			return zone.FeedbackRPE, true
		}
	}
	return 0, false
}

// Label is the short text shown for an RPE, such as "Z2 · Resistência".
func Label(rpe float64) string {
	number := ForRPE(rpe)
	return fmt.Sprintf("Z%d · %s", number, All[number-1].Name)
}
