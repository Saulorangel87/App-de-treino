package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
)

// reevaluationFreeStore faz a reavaliação terminar sem trabalho (perfil
// incompleto), isolando o teste do contrato HTTP do endpoint.
type reevaluationFreeStore struct{ *httpTestPlanStore }

func (reevaluationFreeStore) PlanningContextByUserID(context.Context, string) (planning.Context, error) {
	return planning.Context{}, planning.ErrIncompleteOnboarding
}

func reportRecoveredRequest(t *testing.T, enabled bool) *httptest.ResponseRecorder {
	t.Helper()
	server := &Server{
		auth:     auth.NewService(&httpTestAuthStore{user: auth.User{ID: "user-1"}}, time.Hour),
		planning: planning.NewService(reevaluationFreeStore{&httpTestPlanStore{plan: planning.Plan{ID: "plan-1"}}}, planning.WithProtectionLevels(enabled)),
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/protection/recovered", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "test-session"})
	response := httptest.NewRecorder()
	server.reportRecovered(response, request)
	return response
}

func TestReportRecoveredIsNotFoundWhileProtectionLevelsAreOff(t *testing.T) {
	response := reportRecoveredRequest(t, false)
	if response.Code != http.StatusNotFound {
		t.Fatalf("got %d, want 404: %s", response.Code, response.Body.String())
	}
}

func TestReportRecoveredReturnsThePlanWhenEnabled(t *testing.T) {
	response := reportRecoveredRequest(t, true)
	if response.Code != http.StatusOK {
		t.Fatalf("got %d, want 200: %s", response.Code, response.Body.String())
	}
	var envelope struct {
		Plan struct {
			ID string `json:"id"`
		} `json:"plan"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Plan.ID != "plan-1" {
		t.Fatalf("expected the current plan in the response, got %+v", envelope)
	}
}

func TestWriteWorkoutErrorMapsAFutureWorkoutToConflict(t *testing.T) {
	response := httptest.NewRecorder()
	if !writeWorkoutError(response, planning.ErrWorkoutInFuture) {
		t.Fatal("the error must be handled")
	}
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "workout_in_future") {
		t.Fatalf("got %d %s, want 409 workout_in_future", response.Code, response.Body)
	}
}
