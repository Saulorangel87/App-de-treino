package httpapi

import (
	"errors"
	"net/http"

	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
)

type limitationsRequest struct {
	Limitations []athlete.Limitation `json:"limitations"`
}

type goalsRequest struct {
	Goals []athlete.Goal `json:"goals"`
}

type availabilityRequest struct {
	Availability []athlete.Availability `json:"availability"`
}

func (s *Server) getOnboarding(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	result, err := s.onboarding.Get(r.Context(), user.ID)
	if errors.Is(err, athlete.ErrProfileMissing) {
		writeError(w, http.StatusConflict, "profile_required", "Conclua os dados básicos antes de continuar.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível carregar o onboarding.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"onboarding": result})
}

func (s *Server) putLimitations(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var input limitationsRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	result, err := s.onboarding.SaveLimitations(r.Context(), user.ID, input.Limitations)
	if !s.handleOnboardingError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"limitations": result})
}

func (s *Server) putGoals(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var input goalsRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	result, err := s.onboarding.SaveGoals(r.Context(), user.ID, input.Goals)
	if !s.handleOnboardingError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"goals": result})
}

func (s *Server) putAvailability(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var input availabilityRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	result, err := s.onboarding.SaveAvailability(r.Context(), user.ID, input.Availability)
	if !s.handleOnboardingError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"availability": result})
}

func (s *Server) handleOnboardingError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, athlete.ErrInvalidOnboarding) {
		writeError(w, http.StatusBadRequest, "invalid_onboarding", "Revise as informações desta etapa.")
		return false
	}
	if errors.Is(err, athlete.ErrProfileMissing) {
		writeError(w, http.StatusConflict, "profile_required", "Conclua os dados básicos antes de continuar.")
		return false
	}
	writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível salvar esta etapa.")
	return false
}
