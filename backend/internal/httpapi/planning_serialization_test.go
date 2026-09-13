package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
)

var errHTTPTestUnused = errors.New("unused http test store operation")

type httpTestAuthStore struct {
	user auth.User
}

func (s *httpTestAuthStore) CreateUser(context.Context, string, string, string) (auth.User, error) {
	return auth.User{}, errHTTPTestUnused
}

func (s *httpTestAuthStore) UserByEmail(context.Context, string) (auth.User, error) {
	return auth.User{}, errHTTPTestUnused
}

func (s *httpTestAuthStore) CreateSession(context.Context, string, []byte, time.Time) error {
	return errHTTPTestUnused
}

func (s *httpTestAuthStore) UserBySessionHash(context.Context, []byte) (auth.User, error) {
	return s.user, nil
}

func (s *httpTestAuthStore) DeleteSession(context.Context, []byte) error {
	return errHTTPTestUnused
}

func (s *httpTestAuthStore) CreateEmailToken(context.Context, string, string, []byte, time.Time) error {
	return errHTTPTestUnused
}

func (s *httpTestAuthStore) VerifyEmailToken(context.Context, []byte) (auth.User, error) {
	return auth.User{}, errHTTPTestUnused
}

func (s *httpTestAuthStore) ResetPasswordWithToken(context.Context, []byte, string) error {
	return errHTTPTestUnused
}

type httpTestPlanStore struct {
	plan planning.Plan
}

func (s *httpTestPlanStore) PlanningContextByUserID(context.Context, string) (planning.Context, error) {
	return planning.Context{}, errHTTPTestUnused
}

func (s *httpTestPlanStore) SaveDraftPlan(context.Context, string, planning.Plan) (planning.Plan, error) {
	return planning.Plan{}, errHTTPTestUnused
}

func (s *httpTestPlanStore) CurrentPlanByUserID(context.Context, string) (planning.Plan, error) {
	return s.plan, nil
}

func (s *httpTestPlanStore) ActivatePlanByUserID(context.Context, string, string) error {
	return errHTTPTestUnused
}

func (s *httpTestPlanStore) StartWorkoutByUserID(context.Context, string, string) error {
	return errHTTPTestUnused
}

func (s *httpTestPlanStore) CompleteWorkoutByUserID(context.Context, string, string, planning.CompletionInput) error {
	return errHTTPTestUnused
}

func (s *httpTestPlanStore) CancelWorkoutByUserID(context.Context, string, string) error {
	return errHTTPTestUnused
}

func (s *httpTestPlanStore) MarkWorkoutMissedByUserID(context.Context, string, string) error {
	return errHTTPTestUnused
}

func (s *httpTestPlanStore) ActivitiesByUserID(context.Context, string) ([]planning.Activity, error) {
	return nil, errHTTPTestUnused
}

func TestCurrentPlanSerializesShadowAuditAfterTransaction(t *testing.T) {
	plan := planning.Plan{
		Workouts: []planning.Workout{{
			Explanation: map[string]any{
				"adaptation_shadow": map[string]any{
					"planned_vs_actual": map[string]any{
						"status": "observed",
					},
					"decision_audit": map[string]any{
						"constraints_applied":   []string{"history_query_gate"},
						"missing_data":          []string{"average_power_watts"},
						"used_for_prescription": false,
					},
				},
			},
		}},
	}
	server := &Server{
		auth:     auth.NewService(&httpTestAuthStore{user: auth.User{ID: "user-1"}}, time.Hour),
		planning: planning.NewService(&httpTestPlanStore{plan: plan}),
	}
	request := httptest.NewRequest(http.MethodGet, "/v1/plans/current", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "test-session"})
	response := httptest.NewRecorder()

	server.currentPlan(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("current plan returned %d, want 200: %s", response.Code, response.Body.String())
	}
	var envelope map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("current plan response is not JSON: %v", err)
	}
	returnedPlan, ok := envelope["plan"].(map[string]any)
	if !ok {
		t.Fatalf("plan response = %#v", envelope["plan"])
	}
	workouts, ok := returnedPlan["workouts"].([]any)
	if !ok || len(workouts) != 1 {
		t.Fatalf("workouts response = %#v", returnedPlan["workouts"])
	}
	workout, ok := workouts[0].(map[string]any)
	if !ok {
		t.Fatalf("workout response = %#v", workouts[0])
	}
	explanation, ok := workout["explanation"].(map[string]any)
	if !ok {
		t.Fatalf("explanation response = %#v", workout["explanation"])
	}
	shadow, ok := explanation["adaptation_shadow"].(map[string]any)
	if !ok {
		t.Fatalf("adaptation shadow response = %#v", explanation["adaptation_shadow"])
	}
	audit, ok := shadow["decision_audit"].(map[string]any)
	if !ok {
		t.Fatalf("decision audit response = %#v", shadow["decision_audit"])
	}
	if audit["used_for_prescription"] != false {
		t.Fatalf("decision audit authority = %#v", audit["used_for_prescription"])
	}
	if audit["constraints_applied"].([]any)[0] != "history_query_gate" {
		t.Fatalf("decision audit constraints = %#v", audit["constraints_applied"])
	}
	if shadow["planned_vs_actual"].(map[string]any)["status"] != "observed" {
		t.Fatalf("planned-vs-actual was not serialized: %#v", shadow["planned_vs_actual"])
	}
}
