package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
)

type loggingPlanStore struct {
	httpTestPlanStore
	input  planning.LogWorkoutInput
	today  string
	logged bool
}

func (s *loggingPlanStore) LogWorkoutByUserID(_ context.Context, _, _ string, input planning.LogWorkoutInput, today string) error {
	s.input, s.today, s.logged = input, today, true
	return nil
}

func logWith(t *testing.T, query, body string) (*httptest.ResponseRecorder, *loggingPlanStore) {
	t.Helper()
	store := &loggingPlanStore{}
	server := &Server{
		auth:     auth.NewService(&httpTestAuthStore{user: auth.User{ID: "user-1"}}, time.Hour),
		planning: planning.NewService(store),
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/workouts/9a1eead7-6168-4d50-8c7c-451301e29d85/log"+query, strings.NewReader(body))
	request.SetPathValue("workoutID", "9a1eead7-6168-4d50-8c7c-451301e29d85")
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	response := httptest.NewRecorder()
	server.logWorkout(response, request)
	return response, store
}

func TestLogWorkoutAcceptsAZoneADurationAndADay(t *testing.T) {
	today := time.Now().UTC().Format("2006-01-02")
	response, store := logWith(t, "?date="+today,
		`{`+feedbackFields+`,"actual_zone":3,"duration_minutes":85,"performed_on":"`+today+`","distance_km":40.5,"notes":"  bom pedal  "}`)
	if response.Code != http.StatusOK || !store.logged {
		t.Fatalf("status %d logged=%t: %s", response.Code, store.logged, response.Body)
	}
	if store.input.DurationMinutes != 85 || store.input.PerformedOn != today || store.input.ActualRPE != 6 || store.input.Notes != "bom pedal" {
		t.Fatalf("unexpected input: %+v", store.input)
	}
	if store.input.DistanceKM == nil || *store.input.DistanceKM != 40.5 || store.today != today {
		t.Fatalf("metrics or today lost: %+v today=%s", store.input, store.today)
	}
}

func TestLogWorkoutRejectsBadInput(t *testing.T) {
	today := time.Now().UTC().Format("2006-01-02")
	for name, body := range map[string]string{
		"sem duração":      `{` + feedbackFields + `,"actual_zone":2}`,
		"duração enorme":   `{` + feedbackFields + `,"actual_zone":2,"duration_minutes":9000}`,
		"data futura":      `{` + feedbackFields + `,"actual_zone":2,"duration_minutes":60,"performed_on":"2099-01-01"}`,
		"data inválida":    `{` + feedbackFields + `,"actual_zone":2,"duration_minutes":60,"performed_on":"hoje"}`,
		"sem zona nem RPE": `{` + feedbackFields + `,"duration_minutes":60,"performed_on":"` + today + `"}`,
		"zona inexistente": `{` + feedbackFields + `,"actual_zone":9,"duration_minutes":60}`,
	} {
		response, store := logWith(t, "", body)
		if response.Code != http.StatusBadRequest || store.logged {
			t.Errorf("%s: status %d logged=%t, want 400 and nothing stored", name, response.Code, store.logged)
		}
	}
}

func TestWriteWorkoutErrorMapsAnInvalidLogToBadRequest(t *testing.T) {
	response := httptest.NewRecorder()
	if !writeWorkoutError(response, planning.ErrInvalidLog) {
		t.Fatal("the error must be handled")
	}
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "invalid_workout_log") {
		t.Fatalf("got %d %s, want 400 invalid_workout_log", response.Code, response.Body)
	}
}
