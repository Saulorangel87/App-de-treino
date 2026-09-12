package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
)

type completeWorkoutInput struct {
	CompletionStatus string   `json:"completion_status"`
	PartialReason    string   `json:"partial_reason"`
	ActualRPE        float64  `json:"actual_rpe"`
	Difficulty       string   `json:"difficulty"`
	PainReported     bool     `json:"pain_reported"`
	FatigueAfter     int      `json:"fatigue_after"`
	Notes            string   `json:"notes"`
	DistanceKM       *float64 `json:"distance_km"`
	ElevationGainM   *int     `json:"elevation_gain_m"`
	AveragePowerW    *int     `json:"average_power_watts"`
	AverageHeartRate *int     `json:"average_heart_rate"`
}

func (s *Server) startWorkout(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	plan, err := s.planning.StartWorkout(r.Context(), user.ID, r.PathValue("workoutID"))
	if writeWorkoutError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"plan": plan})
}

func (s *Server) completeWorkout(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var input completeWorkoutInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.CompletionStatus == "" {
		writeWorkoutError(w, planning.ErrInvalidFeedback)
		return
	}
	plan, err := s.planning.CompleteWorkout(r.Context(), user.ID, r.PathValue("workoutID"), planning.CompletionInput{
		CompletionStatus: input.CompletionStatus, PartialReason: strings.TrimSpace(input.PartialReason),
		ActualRPE: input.ActualRPE, Difficulty: input.Difficulty,
		PainReported: input.PainReported, FatigueAfter: input.FatigueAfter,
		Notes:      strings.TrimSpace(input.Notes),
		DistanceKM: input.DistanceKM, ElevationGainM: input.ElevationGainM,
		AveragePowerW: input.AveragePowerW, AverageHeartRate: input.AverageHeartRate,
	})
	if writeWorkoutError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"plan": plan})
}

func (s *Server) cancelWorkout(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	plan, err := s.planning.CancelWorkout(r.Context(), user.ID, r.PathValue("workoutID"))
	if writeWorkoutError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"plan": plan})
}

func (s *Server) markWorkoutMissed(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	plan, err := s.planning.MarkWorkoutMissed(r.Context(), user.ID, r.PathValue("workoutID"))
	if writeWorkoutError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"plan": plan})
}

func writeWorkoutError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, planning.ErrInvalidWorkoutID):
		writeError(w, http.StatusBadRequest, "invalid_workout_id", "O identificador do treino é inválido.")
	case errors.Is(err, planning.ErrInvalidFeedback):
		writeError(w, http.StatusBadRequest, "invalid_feedback", "Informe conclusão completa ou parcial, motivo quando parcial, RPE de 1 a 10, fadiga de 1 a 5 e uma dificuldade válida.")
	case errors.Is(err, planning.ErrWorkoutMissing):
		writeError(w, http.StatusNotFound, "workout_not_found", "O treino não pertence ao seu plano ativo.")
	case errors.Is(err, planning.ErrInvalidTransition):
		writeError(w, http.StatusConflict, "invalid_workout_transition", "O treino não está no estado necessário para esta ação.")
	case errors.Is(err, planning.ErrWorkoutNotPast):
		writeError(w, http.StatusConflict, "workout_not_past", "Esse treino ainda não passou da data planejada.")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível atualizar a sessão.")
	}
	return true
}
