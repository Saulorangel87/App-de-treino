package httpapi

import (
	"errors"
	"net/http"

	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
)

type submaxAssessmentRequest struct {
	DurationMinutes int     `json:"duration_minutes"`
	ActualRPE       float64 `json:"actual_rpe"`
	PainReported    bool    `json:"pain_reported"`
	Notes           string  `json:"notes"`
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
	result, err := s.assessments.SaveSubmax(r.Context(), user.ID, athlete.Assessment{
		DurationMinutes: input.DurationMinutes, ActualRPE: input.ActualRPE, PainReported: input.PainReported, Notes: input.Notes,
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
