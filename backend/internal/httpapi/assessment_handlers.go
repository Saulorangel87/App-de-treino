package httpapi

import (
	"errors"
	"net/http"

	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
	"github.com/Saulorangel87/App-de-treino/backend/internal/zones"
)

type submaxAssessmentRequest struct {
	DurationMinutes int `json:"duration_minutes"`
	// ActualZone (1 to 5) is what the screen sends; the zone wins over ActualRPE,
	// which stays for older clients.
	ActualZone          int      `json:"actual_zone"`
	ActualRPE           float64  `json:"actual_rpe"`
	PainReported        bool     `json:"pain_reported"`
	Notes               string   `json:"notes"`
	AverageHeartRate    *int     `json:"average_heart_rate"`
	AveragePowerW       *int     `json:"average_power_w"`
	DistanceKM          *float64 `json:"distance_km"`
	HeartRateFirstHalf  *int     `json:"heart_rate_first_half"`
	HeartRateSecondHalf *int     `json:"heart_rate_second_half"`
}

// assessmentHistory lists the latest assessments, newest first, with the derived
// efficiency and drift, so the screen can compare one with the previous.
func (s *Server) assessmentHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	history, err := s.assessments.History(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível carregar suas avaliações.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"assessments": history})
}

func (s *Server) currentAssessment(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	result, err := s.assessments.Current(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível carregar sua avaliação.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"assessment": result})
}

func (s *Server) saveSubmaxAssessment(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var input submaxAssessmentRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.ActualZone != 0 {
		rpe, valid := zones.RPEForZone(input.ActualZone)
		if !valid {
			writeError(w, http.StatusBadRequest, "invalid_assessment", "Revise os dados da avaliação.")
			return
		}
		input.ActualRPE = rpe
	}
	result, err := s.assessments.SaveSubmax(r.Context(), user.ID, athlete.Assessment{
		DurationMinutes: input.DurationMinutes, ActualRPE: input.ActualRPE, PainReported: input.PainReported, Notes: input.Notes,
		AverageHeartRate: input.AverageHeartRate, AveragePowerW: input.AveragePowerW, DistanceKM: input.DistanceKM,
		HeartRateFirstHalf: input.HeartRateFirstHalf, HeartRateSecondHalf: input.HeartRateSecondHalf,
	})
	if errors.Is(err, athlete.ErrInvalidAssessment) {
		writeError(w, http.StatusBadRequest, "invalid_assessment", "Revise os dados da avaliação.")
		return
	}
	if errors.Is(err, athlete.ErrProfileMissing) {
		writeError(w, http.StatusConflict, "profile_required", "Conclua seu perfil antes da avaliação.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível salvar sua avaliação.")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"assessment": result})
}
