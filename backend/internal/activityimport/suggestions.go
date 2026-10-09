package activityimport

import (
	"context"
	"math"
	"time"
)

// Limites das sugestões. Uma leitura isolada pode ser erro de sensor (cinta
// solta, pico de cadência), então só se sugere com atividades suficientes.
const (
	minSuggestionActivities = 3
	minPlausibleMaxHR       = 100
	maxPlausibleMaxHR       = 230
	minPlausibleFTP         = 50
	maxPlausibleFTP         = 600
	// ftpFactorFrom20Min é o fator do protocolo clássico de 20 minutos.
	ftpFactorFrom20Min = 0.95
)

// ReferenceSuggestion é um valor sugerido a partir das atividades do próprio
// atleta. Nunca é gravado sozinho: o atleta aceita ou ignora.
type ReferenceSuggestion struct {
	Value      int    `json:"value"`
	Activities int    `json:"activities"`
	ObservedOn string `json:"observed_on"`
}

// ReferenceSuggestions reúne as sugestões de frequência máxima e de FTP.
type ReferenceSuggestions struct {
	MaxHeartRate *ReferenceSuggestion `json:"max_heart_rate,omitempty"`
	FTP          *ReferenceSuggestion `json:"ftp,omitempty"`
}

// ReferenceSuggestions sugere a frequência máxima (maior batimento visto) e o
// FTP (95% da melhor média de 20 minutos) com base nas atividades importadas.
func (s *Service) ReferenceSuggestions(ctx context.Context, userID string) (ReferenceSuggestions, error) {
	activities, err := s.store.ListActivities(ctx, userID)
	if err != nil {
		return ReferenceSuggestions{}, err
	}
	return suggestionsFrom(activities), nil
}

func suggestionsFrom(activities []Activity) ReferenceSuggestions {
	var suggestions ReferenceSuggestions
	var hrCount, powerCount int
	var bestHR, bestPower int
	var hrAt, powerAt time.Time
	for _, activity := range activities {
		if hr := activity.MaxHeartRate; hr != nil && *hr >= minPlausibleMaxHR && *hr <= maxPlausibleMaxHR {
			hrCount++
			if *hr > bestHR {
				bestHR, hrAt = *hr, activity.StartedAt
			}
		}
		if power := activity.Best20MinPowerW; power != nil && *power > 0 {
			powerCount++
			if *power > bestPower {
				bestPower, powerAt = *power, activity.StartedAt
			}
		}
	}
	if hrCount >= minSuggestionActivities {
		suggestions.MaxHeartRate = &ReferenceSuggestion{Value: bestHR, Activities: hrCount, ObservedOn: hrAt.Format("2006-01-02")}
	}
	if powerCount >= minSuggestionActivities {
		ftp := int(math.Round(float64(bestPower) * ftpFactorFrom20Min))
		if ftp >= minPlausibleFTP && ftp <= maxPlausibleFTP {
			suggestions.FTP = &ReferenceSuggestion{Value: ftp, Activities: powerCount, ObservedOn: powerAt.Format("2006-01-02")}
		}
	}
	return suggestions
}
