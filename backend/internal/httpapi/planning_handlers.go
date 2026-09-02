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
	if !user.EmailVerified {
		writeError(w, http.StatusForbidden, "email_not_verified", "Confirme seu e-mail antes de gerar um plano.")
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

func (s *Server) activatePlan(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	if !user.EmailVerified {
		writeError(w, http.StatusForbidden, "email_not_verified", "Confirme seu e-mail antes de ativar um plano.")
		return
	}
	plan, err := s.planning.Activate(r.Context(), user.ID, r.PathValue("planID"))
	if errors.Is(err, planning.ErrInvalidPlanID) {
		writeError(w, http.StatusBadRequest, "invalid_plan_id", "O identificador do plano é inválido.")
		return
	}
	if errors.Is(err, planning.ErrPlanMissing) {
		writeError(w, http.StatusNotFound, "plan_not_found", "O rascunho não foi encontrado ou já foi ativado.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível ativar o plano.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"plan": plan})
}
