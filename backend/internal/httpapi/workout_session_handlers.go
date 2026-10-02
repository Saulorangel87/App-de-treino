package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
	"github.com/Saulorangel87/App-de-treino/backend/internal/zones"
)

type completeWorkoutInput struct {
	CompletionStatus string  `json:"completion_status"`
	PartialReason    string  `json:"partial_reason"`
	ActualRPE        float64 `json:"actual_rpe"`
	// ActualZone (1 to 5) is the zone the athlete rode in; the app sends it instead
	// of actual_rpe. It is stored as the RPE that represents the zone.
	ActualZone         int      `json:"actual_zone"`
	Difficulty         string   `json:"difficulty"`
	PainReported       bool     `json:"pain_reported"`
	FatigueAfter       int      `json:"fatigue_after"`
	RecoveryAfter      *int     `json:"recovery_after"`
	RepeatConfidence   *int     `json:"repeat_confidence"`
	Satisfaction       *int     `json:"satisfaction"`
	Terrain            string   `json:"terrain"`
	ExternalConditions string   `json:"external_conditions"`
	EquipmentUsed      string   `json:"equipment_used"`
	Notes              string   `json:"notes"`
	DistanceKM         *float64 `json:"distance_km"`
	ElevationGainM     *int     `json:"elevation_gain_m"`
	AveragePowerW      *int     `json:"average_power_watts"`
	AverageHeartRate   *int     `json:"average_heart_rate"`
	AverageCadenceRPM  *int     `json:"average_cadence_rpm"`
}

// logWorkoutInput is the feedback of a ride the athlete did without the
// stopwatch, plus how long it took and on which day.
type logWorkoutInput struct {
	completeWorkoutInput
	DurationMinutes int    `json:"duration_minutes"`
	PerformedOn     string `json:"performed_on"`
}

type correctWorkoutInput struct {
	DistanceKM        *float64 `json:"distance_km"`
	ElevationGainM    *int     `json:"elevation_gain_m"`
	AveragePowerW     *int     `json:"average_power_watts"`
	AverageHeartRate  *int     `json:"average_heart_rate"`
	AverageCadenceRPM *int     `json:"average_cadence_rpm"`
}

func (s *Server) startWorkout(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	plan, err := s.planning.StartWorkout(r.Context(), user.ID, r.PathValue("workoutID"), r.URL.Query().Get("date"))
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
	completion, ok := completionInputFrom(w, input)
	if !ok {
		return
	}
	plan, err := s.planning.CompleteWorkout(r.Context(), user.ID, r.PathValue("workoutID"), completion)
	if writeWorkoutError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"plan": plan})
}

// completionInputFrom checks the feedback fields shared by completing a session
// and logging a ride, turns a reported zone into its RPE and trims the free text.
// It writes the error response and returns false when the input is invalid.
func completionInputFrom(w http.ResponseWriter, input completeWorkoutInput) (planning.CompletionInput, bool) {
	if input.CompletionStatus == "" {
		writeWorkoutError(w, planning.ErrInvalidFeedback)
		return planning.CompletionInput{}, false
	}
	if input.ActualZone != 0 {
		rpe, ok := zones.RPEForZone(input.ActualZone)
		if !ok {
			writeWorkoutError(w, planning.ErrInvalidFeedback)
			return planning.CompletionInput{}, false
		}
		input.ActualRPE = rpe
	}
	return planning.CompletionInput{
		CompletionStatus: input.CompletionStatus, PartialReason: strings.TrimSpace(input.PartialReason),
		ActualRPE: input.ActualRPE, Difficulty: input.Difficulty,
		PainReported: input.PainReported, FatigueAfter: input.FatigueAfter,
		RecoveryAfter: input.RecoveryAfter, RepeatConfidence: input.RepeatConfidence,
		Satisfaction: input.Satisfaction, Terrain: strings.TrimSpace(input.Terrain),
		ExternalConditions: strings.TrimSpace(input.ExternalConditions),
		EquipmentUsed:      strings.TrimSpace(input.EquipmentUsed),
		Notes:              strings.TrimSpace(input.Notes),
		DistanceKM:         input.DistanceKM, ElevationGainM: input.ElevationGainM,
		AveragePowerW: input.AveragePowerW, AverageHeartRate: input.AverageHeartRate, AverageCadenceRPM: input.AverageCadenceRPM,
	}, true
}

// logWorkout marks a planned workout as done without the stopwatch ("task mode").
func (s *Server) logWorkout(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var input logWorkoutInput
	if !decodeJSON(w, r, &input) {
		return
	}
	completion, ok := completionInputFrom(w, input.completeWorkoutInput)
	if !ok {
		return
	}
	plan, err := s.planning.LogWorkout(r.Context(), user.ID, r.PathValue("workoutID"), planning.LogWorkoutInput{
		CompletionInput: completion, DurationMinutes: input.DurationMinutes, PerformedOn: strings.TrimSpace(input.PerformedOn),
	}, r.URL.Query().Get("date"))
	if writeWorkoutError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"plan": plan})
}

func (s *Server) correctWorkout(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var input correctWorkoutInput
	if !decodeJSON(w, r, &input) {
		return
	}
	plan, err := s.planning.CorrectWorkout(r.Context(), user.ID, r.PathValue("workoutID"), planning.WorkoutCorrectionInput{
		DistanceKM: input.DistanceKM, ElevationGainM: input.ElevationGainM,
		AveragePowerW: input.AveragePowerW, AverageHeartRate: input.AverageHeartRate, AverageCadenceRPM: input.AverageCadenceRPM,
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

func (s *Server) undoWorkout(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	plan, err := s.planning.UndoWorkout(r.Context(), user.ID, r.PathValue("workoutID"))
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
		writeError(w, http.StatusBadRequest, "invalid_feedback", "Informe conclusão completa ou parcial, motivo quando parcial, a zona de esforço (1 a 5), fadiga de 1 a 5 e uma dificuldade válida.")
	case errors.Is(err, planning.ErrInvalidCorrection):
		writeError(w, http.StatusBadRequest, "invalid_workout_correction", "Informe métricas do pedal válidas ou remova o valor que não deseja manter.")
	case errors.Is(err, planning.ErrWorkoutCorrection):
		writeError(w, http.StatusConflict, "workout_correction_not_allowed", "Somente uma sessão concluída com dados inelegíveis pode receber correção.")
	case errors.Is(err, planning.ErrWorkoutSafetyBlocked):
		writeError(w, http.StatusConflict, "workout_blocked_by_safety", "Este treino foi bloqueado porque existe uma limitação ativa. Gere um novo plano protegido ou procure orientação profissional antes de retomar a intensidade.")
	case errors.Is(err, planning.ErrWorkoutMissing):
		writeError(w, http.StatusNotFound, "workout_not_found", "O treino não pertence ao seu plano ativo.")
	case errors.Is(err, planning.ErrInvalidTransition):
		writeError(w, http.StatusConflict, "invalid_workout_transition", "O treino não está no estado necessário para esta ação.")
	case errors.Is(err, planning.ErrInvalidLog):
		writeError(w, http.StatusBadRequest, "invalid_workout_log", "Informe a duração entre 1 e 720 minutos e um dia entre hoje e os últimos 7 dias, sem ser anterior ao dia planejado do treino.")
	case errors.Is(err, planning.ErrWorkoutInFuture):
		writeError(w, http.StatusConflict, "workout_in_future", "Este treino é de uma data futura. Ele fica disponível no dia planejado.")
	case errors.Is(err, planning.ErrWorkoutNotPast):
		writeError(w, http.StatusConflict, "workout_not_past", "Esse treino ainda não passou da data planejada.")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível atualizar a sessão.")
	}
	return true
}
