package httpapi

import (
	"errors"
	"net/http"

	"github.com/Saulorangel87/App-de-treino/backend/internal/feedback"
)

func (s *Server) createUserFeedback(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var input feedback.CreateInput
	if !decodeJSON(w, r, &input) {
		return
	}
	entry, err := s.feedback.Create(r.Context(), user.ID, input)
	if errors.Is(err, feedback.ErrInvalidInput) {
		writeError(w, http.StatusBadRequest, "invalid_feedback", "Escolha uma categoria, uma nota de 1 a 5 e escreva pelo menos 10 caracteres.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível registrar seu feedback.")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"feedback": entry, "message": "Obrigado por compartilhar sua experiência."})
}
