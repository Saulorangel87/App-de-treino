package httpapi

import (
	"errors"
	"net/http"

	"github.com/Saulorangel87/App-de-treino/backend/internal/ai"
	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
)

func (s *Server) explainWorkout(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	workout, err := s.planning.Workout(r.Context(), user.ID, r.PathValue("workoutID"))
	if errors.Is(err, planning.ErrInvalidWorkoutID) {
		writeError(w, http.StatusBadRequest, "invalid_workout_id", "O identificador do treino é inválido.")
		return
	}
	if errors.Is(err, planning.ErrWorkoutMissing) || errors.Is(err, planning.ErrPlanMissing) {
		writeError(w, http.StatusNotFound, "workout_not_found", "O treino não foi encontrado no seu plano.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível carregar o treino.")
		return
	}

	fallback := workoutExplanationSummary(workout)
	input := ai.ExplanationInput{
		WorkoutName: workout.Name, Objective: workout.Objective,
		DurationMinutes: workout.DurationMinutes, TargetRPE: workout.TargetRPE,
		Rules:         workoutExplanationRules(workout),
		EvidenceScope: workoutExplanationEvidenceScope(workout),
	}
	if s.ai == nil || !s.ai.Enabled() {
		writeJSON(w, http.StatusOK, map[string]any{
			"explanation": fallback, "source": "rules", "ai_enabled": false,
		})
		return
	}
	text, err := s.ai.Explain(r.Context(), input)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"explanation": fallback, "source": "rules_fallback", "ai_enabled": true,
			"warning": "A explicação automática está indisponível; exibimos a explicação validada do motor.",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"explanation": text, "source": "ollama", "ai_enabled": true,
	})
}

func workoutExplanationSummary(workout planning.Workout) string {
	if summary, ok := workout.Explanation["summary"].(string); ok && summary != "" {
		return summary
	}
	return "Esta sessão foi escolhida pelas regras do seu plano, respeitando seu objetivo, experiência e tempo disponível."
}

func workoutExplanationRules(workout planning.Workout) []string {
	values, ok := workout.Explanation["rules"].([]any)
	if ok {
		rules := make([]string, 0, len(values))
		for _, value := range values {
			if rule, ok := value.(string); ok {
				rules = append(rules, rule)
			}
		}
		return rules
	}
	if values, ok := workout.Explanation["rules"].([]string); ok {
		return append([]string(nil), values...)
	}
	return nil
}

func workoutExplanationEvidenceScope(workout planning.Workout) string {
	scope, _ := workout.Explanation["evidence_scope"].(string)
	return scope
}
