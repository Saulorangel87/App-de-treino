package httpapi

import "net/http"

func (s *Server) activities(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	activities, err := s.planning.Activities(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível carregar as atividades.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"activities": activities})
}
