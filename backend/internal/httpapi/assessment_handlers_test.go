package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
)

type recordingAssessmentStore struct {
	saved   athlete.Assessment
	history []athlete.Assessment
}

func (s *recordingAssessmentStore) CurrentAssessmentByUserID(context.Context, string) (*athlete.Assessment, error) {
	return nil, nil
}
func (s *recordingAssessmentStore) AssessmentHistoryByUserID(context.Context, string, int) ([]athlete.Assessment, error) {
	return s.history, nil
}
func (s *recordingAssessmentStore) SaveSubmaxAssessment(_ context.Context, _ string, input athlete.Assessment) (athlete.Assessment, error) {
	s.saved = input
	return input, nil
}

func assessmentServer(store *recordingAssessmentStore) *Server {
	return &Server{
		auth:        auth.NewService(&httpTestAuthStore{user: auth.User{ID: "user-1"}}, time.Hour),
		assessments: athlete.NewAssessmentService(store),
	}
}

func postAssessment(t *testing.T, store *recordingAssessmentStore, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/v1/assessments/submaximal", strings.NewReader(body))
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	response := httptest.NewRecorder()
	assessmentServer(store).saveSubmaxAssessment(response, request)
	return response
}

func TestAssessmentAcceptsAZoneAndTheNumbers(t *testing.T) {
	store := &recordingAssessmentStore{}
	response := postAssessment(t, store,
		`{"duration_minutes":20,"actual_zone":2,"pain_reported":false,"average_heart_rate":140,"average_power_w":168,"heart_rate_first_half":135,"heart_rate_second_half":144}`)
	if response.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", response.Code, response.Body)
	}
	if store.saved.ActualRPE != 4.5 || !store.saved.EligibleForProgression {
		t.Fatalf("zone 2 must be stored as RPE 4.5 and be eligible: %+v", store.saved)
	}
	if store.saved.AverageHeartRate == nil || *store.saved.AverageHeartRate != 140 {
		t.Fatalf("heart rate lost: %+v", store.saved)
	}
	if !strings.Contains(response.Body.String(), `"heart_rate_drift_percent":6.7`) || !strings.Contains(response.Body.String(), `"power_per_bpm"`) {
		t.Fatalf("derived numbers missing from the response: %s", response.Body)
	}
}

func TestAssessmentZoneWinsOverRPEAndOldClientsStillWork(t *testing.T) {
	store := &recordingAssessmentStore{}
	if r := postAssessment(t, store, `{"duration_minutes":20,"actual_zone":4,"actual_rpe":3}`); r.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", r.Code, r.Body)
	}
	if store.saved.ActualRPE != 7 || store.saved.EligibleForProgression {
		t.Fatalf("zone 4 must win over the RPE and not be eligible: %+v", store.saved)
	}
	if r := postAssessment(t, store, `{"duration_minutes":20,"actual_rpe":5}`); r.Code != http.StatusCreated || store.saved.ActualRPE != 5 {
		t.Fatalf("a client that still sends actual_rpe must keep working: %d %+v", r.Code, store.saved)
	}
}

func TestAssessmentRejectsBadInput(t *testing.T) {
	for name, body := range map[string]string{
		"zona inexistente": `{"duration_minutes":20,"actual_zone":9}`,
		"sem zona nem RPE": `{"duration_minutes":20}`,
		"uma metade só":    `{"duration_minutes":20,"actual_zone":2,"heart_rate_first_half":130}`,
		"FC absurda":       `{"duration_minutes":20,"actual_zone":2,"average_heart_rate":400}`,
	} {
		store := &recordingAssessmentStore{}
		if r := postAssessment(t, store, body); r.Code != http.StatusBadRequest || store.saved.DurationMinutes != 0 {
			t.Errorf("%s: status %d saved=%+v, want 400 and nothing stored", name, r.Code, store.saved)
		}
	}
}

func TestAssessmentHistoryDerivesTheNumbers(t *testing.T) {
	hr, power := 140, 168
	store := &recordingAssessmentStore{history: []athlete.Assessment{{ID: "a", DurationMinutes: 20, AverageHeartRate: &hr, AveragePowerW: &power}}}
	request := httptest.NewRequest(http.MethodGet, "/v1/assessments", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	response := httptest.NewRecorder()
	assessmentServer(store).assessmentHistory(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"power_per_bpm"`) || !strings.Contains(response.Body.String(), `"value":1.2`) {
		t.Fatalf("status %d: %s", response.Code, response.Body)
	}
}
