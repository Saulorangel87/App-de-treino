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

// recordingPlanStore captures what the handler sends to the store on completion.
type recordingPlanStore struct {
	httpTestPlanStore
	completion planning.CompletionInput
	completed  bool
}

func (s *recordingPlanStore) CompleteWorkoutByUserID(_ context.Context, _, _ string, input planning.CompletionInput) error {
	s.completion, s.completed = input, true
	return nil
}

func completeWith(t *testing.T, body string) (*httptest.ResponseRecorder, *recordingPlanStore) {
	t.Helper()
	store := &recordingPlanStore{}
	server := &Server{
		auth:     auth.NewService(&httpTestAuthStore{user: auth.User{ID: "user-1"}}, time.Hour),
		planning: planning.NewService(store),
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/workouts/9a1eead7-6168-4d50-8c7c-451301e29d85/complete", strings.NewReader(body))
	request.SetPathValue("workoutID", "9a1eead7-6168-4d50-8c7c-451301e29d85")
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	response := httptest.NewRecorder()
	server.completeWorkout(response, request)
	return response, store
}

const feedbackFields = `"completion_status":"complete","difficulty":"moderate","fatigue_after":3`

func TestCompleteWorkoutStoresAReportedZoneAsItsRPE(t *testing.T) {
	for zone, want := range map[int]float64{1: 3, 2: 4.5, 3: 6, 4: 7, 5: 8.5} {
		response, store := completeWith(t, `{`+feedbackFields+`,"actual_zone":`+string(rune('0'+zone))+`}`)
		if response.Code != http.StatusOK || !store.completed || store.completion.ActualRPE != want {
			t.Errorf("zone %d: status %d, stored RPE %v, want %v", zone, response.Code, store.completion.ActualRPE, want)
		}
	}
}

func TestCompleteWorkoutStillAcceptsRPEForOlderClients(t *testing.T) {
	response, store := completeWith(t, `{`+feedbackFields+`,"actual_rpe":7}`)
	if response.Code != http.StatusOK || store.completion.ActualRPE != 7 {
		t.Fatalf("status %d, stored RPE %v, want 200 and 7", response.Code, store.completion.ActualRPE)
	}
}

func TestCompleteWorkoutRejectsAnUnknownZoneAndAMissingEffort(t *testing.T) {
	for name, body := range map[string]string{
		"zona 0 sem RPE": `{` + feedbackFields + `}`,
		"zona 6":         `{` + feedbackFields + `,"actual_zone":6}`,
		"zona negativa":  `{` + feedbackFields + `,"actual_zone":-2}`,
		"zona 6 com RPE": `{` + feedbackFields + `,"actual_zone":6,"actual_rpe":5}`,
	} {
		response, store := completeWith(t, body)
		if response.Code != http.StatusBadRequest || store.completed {
			t.Errorf("%s: status %d completed=%t, want 400 and nothing stored", name, response.Code, store.completed)
		}
	}
}

func TestZoneWinsOverRPEWhenBothAreSent(t *testing.T) {
	_, store := completeWith(t, `{`+feedbackFields+`,"actual_zone":2,"actual_rpe":9}`)
	if store.completion.ActualRPE != 4.5 {
		t.Fatalf("the zone must decide the stored RPE, got %v", store.completion.ActualRPE)
	}
}
