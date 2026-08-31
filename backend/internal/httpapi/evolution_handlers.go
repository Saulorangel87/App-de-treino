package httpapi

import "net/http"

func (s *Server) evolutionSummary(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	summary, err := s.evolution.Summary(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível carregar sua evolução.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"summary": summary})
}
