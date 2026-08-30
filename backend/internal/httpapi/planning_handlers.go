package httpapi

import (
	"errors"
	"net/http"

	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
)

func (s *Server) generatePlan(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	plan, err := s.planning.Generate(r.Context(), user.ID)
	if errors.Is(err, planning.ErrIncompleteOnboarding) {
		writeError(w, http.StatusConflict, "onboarding_incomplete", "Conclua perfil, objetivo e disponibilidade antes de gerar o plano.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível gerar o plano.")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"plan": plan})
}

func (s *Server) currentPlan(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	plan, err := s.planning.Current(r.Context(), user.ID)
	if errors.Is(err, planning.ErrPlanMissing) {
		writeJSON(w, http.StatusOK, map[string]any{"plan": nil})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível carregar o plano.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"plan": plan})
}
