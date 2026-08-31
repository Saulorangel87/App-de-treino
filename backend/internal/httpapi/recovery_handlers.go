package httpapi

import (
	"errors"
	"net/http"

	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
)

type recoveryRequest struct {
	RecordedOn   string `json:"recorded_on"`
	SleepMinutes int    `json:"sleep_minutes"`
	SleepQuality int    `json:"sleep_quality"`
	StressLevel  int    `json:"stress_level"`
	FatigueLevel int    `json:"fatigue_level"`
	Notes        string `json:"notes"`
}

func (s *Server) todayRecovery(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	recordedOn := r.URL.Query().Get("date")
	result, err := s.recovery.Get(r.Context(), user.ID, recordedOn)
	if errors.Is(err, athlete.ErrInvalidRecovery) {
		writeError(w, http.StatusBadRequest, "invalid_date", "Informe a data local no formato AAAA-MM-DD.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível carregar seu check-in.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"recovery": result})
}

func (s *Server) putTodayRecovery(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var input recoveryRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	result, err := s.recovery.Save(r.Context(), user.ID, athlete.RecoveryCheckin{
		RecordedOn: input.RecordedOn, SleepMinutes: input.SleepMinutes, SleepQuality: input.SleepQuality,
		StressLevel: input.StressLevel, FatigueLevel: input.FatigueLevel, Notes: input.Notes,
	})
	if errors.Is(err, athlete.ErrInvalidRecovery) {
		writeError(w, http.StatusBadRequest, "invalid_recovery", "Revise os dados do check-in diário.")
		return
	}
	if errors.Is(err, athlete.ErrProfileMissing) {
		writeError(w, http.StatusConflict, "profile_required", "Conclua seu perfil antes do check-in.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível salvar seu check-in.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"recovery": result})
}
