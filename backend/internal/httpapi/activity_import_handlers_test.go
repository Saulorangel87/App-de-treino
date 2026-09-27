package httpapi

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/activityimport"
	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
)

const sampleImportGPX = `<?xml version="1.0" encoding="UTF-8"?>
<gpx><trk><type>cycling</type><trkseg>
<trkpt lat="-23.5" lon="-46.6"><ele>760</ele><time>2026-09-20T08:00:00Z</time></trkpt>
<trkpt lat="-23.51" lon="-46.61"><ele>780</ele><time>2026-09-20T08:10:00Z</time></trkpt>
</trkseg></trk></gpx>`

type fakeActivityImportStore struct {
	hashes  map[string]bool
	deleted string
}

func (f *fakeActivityImportStore) ActivityExists(_ context.Context, _, fileHash string) (bool, error) {
	return f.hashes[fileHash], nil
}

func (f *fakeActivityImportStore) SaveActivity(_ context.Context, userID, source, fileHash string, parsed activityimport.Parsed) (activityimport.Activity, error) {
	f.hashes[fileHash] = true
	return activityimport.Activity{ID: "activity-1", UserID: userID, Source: source, Parsed: parsed, ImportedAt: time.Now()}, nil
}

func (f *fakeActivityImportStore) ListActivities(_ context.Context, userID string) ([]activityimport.Activity, error) {
	return []activityimport.Activity{{ID: "activity-1", UserID: userID, Source: activityimport.SourceGPX}}, nil
}

func (f *fakeActivityImportStore) DeleteActivity(_ context.Context, _, activityID string) error {
	if activityID != "activity-1" {
		return activityimport.ErrNotFound
	}
	f.deleted = activityID
	return nil
}

func (f *fakeActivityImportStore) WorkoutCandidatesOnDate(_ context.Context, _ string, _ time.Time) ([]activityimport.WorkoutCandidate, error) {
	return []activityimport.WorkoutCandidate{{ID: "workout-1", ScheduledOn: "2026-09-20", Name: "Sweet spot", Status: "planned"}}, nil
}

func newActivityImportTestServer() (*Server, *fakeActivityImportStore) {
	store := &fakeActivityImportStore{hashes: map[string]bool{}}
	authStore := &httpTestAuthStore{user: auth.User{ID: "user-1"}}
	server := &Server{auth: auth.NewService(authStore, time.Hour), activityImport: activityimport.NewService(store)}
	return server, store
}

func multipartGPXRequest(t *testing.T, filename, content string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/activities/import", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	return req
}

func TestImportActivity_Success(t *testing.T) {
	server, _ := newActivityImportTestServer()
	req := multipartGPXRequest(t, "ride.gpx", sampleImportGPX)
	response := httptest.NewRecorder()
	server.importActivity(response, req)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("candidate_workouts")) {
		t.Errorf("response should include candidate_workouts: %s", response.Body.String())
	}
}

func TestImportActivity_RequiresAuth(t *testing.T) {
	server, _ := newActivityImportTestServer()
	req := httptest.NewRequest(http.MethodPost, "/v1/activities/import", nil)
	response := httptest.NewRecorder()
	server.importActivity(response, req)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestImportActivity_RejectsUnrecognizedFormat(t *testing.T) {
	server, _ := newActivityImportTestServer()
	req := multipartGPXRequest(t, "ride.tcx", sampleImportGPX)
	response := httptest.NewRecorder()
	server.importActivity(response, req)
	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want 415; body: %s", response.Code, response.Body.String())
	}
}

func TestImportActivity_MissingFileField(t *testing.T) {
	server, _ := newActivityImportTestServer()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/v1/activities/import", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	response := httptest.NewRecorder()
	server.importActivity(response, req)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", response.Code, response.Body.String())
	}
}

func TestListImportedActivities(t *testing.T) {
	server, _ := newActivityImportTestServer()
	req := httptest.NewRequest(http.MethodGet, "/v1/activities/imported", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	response := httptest.NewRecorder()
	server.listImportedActivities(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", response.Code, response.Body.String())
	}
}

func TestDeleteImportedActivity_NotFound(t *testing.T) {
	server, _ := newActivityImportTestServer()
	req := httptest.NewRequest(http.MethodDelete, "/v1/activities/imported/unknown-id", nil)
	req.SetPathValue("activityID", "unknown-id")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	response := httptest.NewRecorder()
	server.deleteImportedActivity(response, req)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body: %s", response.Code, response.Body.String())
	}
}

func TestDeleteImportedActivity_Success(t *testing.T) {
	server, store := newActivityImportTestServer()
	req := httptest.NewRequest(http.MethodDelete, "/v1/activities/imported/activity-1", nil)
	req.SetPathValue("activityID", "activity-1")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	response := httptest.NewRecorder()
	server.deleteImportedActivity(response, req)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body: %s", response.Code, response.Body.String())
	}
	if store.deleted != "activity-1" {
		t.Errorf("store.deleted = %q, want activity-1", store.deleted)
	}
}
