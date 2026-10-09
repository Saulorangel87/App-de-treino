package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Saulorangel87/App-de-treino/backend/internal/activityimport"
)

// maxImportRequestBytes bounds the whole multipart request, not just the
// file field, so a malicious Content-Length can't force reading past the
// documented limit before activityimport.Service even sees the data.
const maxImportRequestBytes = activityimport.MaxFileSizeBytes + (1 << 20)

func (s *Server) importActivity(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxImportRequestBytes)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "file_too_large", "O arquivo excede o limite de 20 MB.")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_upload", "Não foi possível ler o arquivo enviado.")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing_file", "Envie um arquivo no campo \"file\".")
		return
	}
	defer file.Close()

	activity, candidates, err := s.activityImport.Import(r.Context(), user.ID, header.Filename, file)
	if writeActivityImportError(w, err) {
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"activity": activity, "candidate_workouts": candidates})
}

func (s *Server) listImportedActivities(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	activities, err := s.activityImport.List(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível carregar as atividades importadas.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"activities": activities})
}

func (s *Server) referenceSuggestions(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	suggestions, err := s.activityImport.ReferenceSuggestions(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível carregar as sugestões.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"suggestions": suggestions})
}

func (s *Server) deleteImportedActivity(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	err := s.activityImport.Delete(r.Context(), user.ID, r.PathValue("activityID"))
	if writeActivityImportError(w, err) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type linkImportedActivityRequest struct {
	// WorkoutID nulo remove o vínculo.
	WorkoutID *string `json:"workout_id"`
}

// linkImportedActivity liga a atividade a um treino do atleta (ou remove o
// vínculo). É só uma associação: nada é copiado para o treino nem muda no plano.
func (s *Server) linkImportedActivity(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var body linkImportedActivityRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Informe o treino no campo \"workout_id\" (ou null para desvincular).")
		return
	}
	activity, err := s.activityImport.Link(r.Context(), user.ID, r.PathValue("activityID"), body.WorkoutID)
	if writeActivityImportError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"activity": activity})
}

func (s *Server) importedActivityCandidates(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	candidates, err := s.activityImport.Candidates(r.Context(), user.ID, r.PathValue("activityID"))
	if writeActivityImportError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"candidate_workouts": candidates})
}

func writeActivityImportError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, activityimport.ErrFileTooLarge):
		writeError(w, http.StatusRequestEntityTooLarge, "file_too_large", "O arquivo excede o limite de 20 MB.")
	case errors.Is(err, activityimport.ErrUnrecognizedFormat):
		writeError(w, http.StatusUnsupportedMediaType, "unrecognized_format", "Envie um arquivo .fit ou .gpx.")
	case errors.Is(err, activityimport.ErrUnsupportedSport):
		writeError(w, http.StatusUnprocessableEntity, "unsupported_sport", "Este arquivo não parece ser de uma atividade de ciclismo.")
	case errors.Is(err, activityimport.ErrNoData):
		writeError(w, http.StatusUnprocessableEntity, "no_activity_data", "Não foi possível encontrar dados de atividade neste arquivo.")
	case errors.Is(err, activityimport.ErrCorruptFile):
		writeError(w, http.StatusUnprocessableEntity, "corrupt_file", "Não foi possível ler este arquivo.")
	case errors.Is(err, activityimport.ErrDuplicate):
		writeError(w, http.StatusConflict, "duplicate_activity", "Este arquivo já foi importado.")
	case errors.Is(err, activityimport.ErrNotFound):
		writeError(w, http.StatusNotFound, "activity_not_found", "Atividade importada não encontrada.")
	case errors.Is(err, activityimport.ErrWorkoutNotFound):
		writeError(w, http.StatusNotFound, "workout_not_found", "Treino não encontrado.")
	case errors.Is(err, activityimport.ErrInvalidWorkoutID):
		writeError(w, http.StatusBadRequest, "invalid_workout_id", "Identificador de treino inválido.")
	case errors.Is(err, io.ErrUnexpectedEOF):
		writeError(w, http.StatusBadRequest, "invalid_upload", "O arquivo enviado está incompleto.")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível processar o arquivo.")
	}
	return true
}
