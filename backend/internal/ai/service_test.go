package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDisabledServiceDoesNotCallProvider(t *testing.T) {
	if _, err := NewService(nil).Explain(context.Background(), ExplanationInput{}); err != ErrDisabled {
		t.Fatalf("expected disabled error, got %v", err)
	}
}

func TestOllamaClientSendsBoundedPortugueseExplanationRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" || r.Method != http.MethodPost {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		var request ollamaChatRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if request.Stream || request.Model != "qwen3:4b-instruct" || request.Options["num_predict"] != float64(180) {
			http.Error(w, "unexpected model options", http.StatusBadRequest)
			return
		}
		if !strings.Contains(request.Messages[0].Content, "não invente estudos") || !strings.Contains(request.Messages[1].Content, "Giro de base") {
			http.Error(w, "safety prompt missing", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":{"content":"Este giro mantém o esforço confortável e respeita o tempo disponível."}}`))
	}))
	defer server.Close()

	client, err := NewOllamaClient(server.URL, "qwen3:4b-instruct", 2*time.Second, 180, 1)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}
	text, err := client.Explain(context.Background(), ExplanationInput{
		WorkoutName: "Giro de base", Objective: "Elevar a resistência", DurationMinutes: 45,
		TargetRPE: 4, Rules: []string{"Carga compatível com sua experiência."}, EvidenceScope: "Progressão gradual.",
	})
	if err != nil || text == "" {
		t.Fatalf("expected explanation, got %q, %v", text, err)
	}
}

func TestOllamaClientRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":{"content":"` + strings.Repeat("x", 1300) + `"}}`))
	}))
	defer server.Close()
	client, err := NewOllamaClient(server.URL, "qwen3:4b-instruct", 2*time.Second, 180, 1)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}
	_, err = client.Explain(context.Background(), ExplanationInput{WorkoutName: "Giro", Objective: "Resistência", DurationMinutes: 30, TargetRPE: 4})
	if err != ErrInvalidResponse {
		t.Fatalf("expected invalid response, got %v", err)
	}
}

func TestWorkerClientSendsOnlyValidatedFacts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/cadencia/explanation" || r.Header.Get("Authorization") != "Bearer worker-secret" {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		var request workerExplanationRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if request.WorkoutName != "Giro de base" || request.Duration != 45 || request.TargetRPE != 4 || len(request.Rules) != 1 {
			http.Error(w, "unexpected payload", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"explanation":"Este giro mantém o esforço confortável."}`))
	}))
	defer server.Close()

	client, err := NewWorkerClient(server.URL, "worker-secret", 2*time.Second, 1)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}
	text, err := client.Explain(context.Background(), ExplanationInput{
		WorkoutName: "Giro de base", Objective: "Elevar a resistência", DurationMinutes: 45,
		TargetRPE: 4, Rules: []string{"Carga compatível."}, EvidenceScope: "Progressão gradual.",
	})
	if err != nil || text == "" {
		t.Fatalf("expected explanation, got %q, %v", text, err)
	}
}

type fakeProvider struct {
	text string
	err  error
}

func (p fakeProvider) Explain(context.Context, ExplanationInput) (string, error) {
	return p.text, p.err
}

func TestFallbackProviderUsesSecondaryOnlyAfterPrimaryFailure(t *testing.T) {
	provider := NewFallbackProvider(
		fakeProvider{err: errors.New("ollama offline")},
		fakeProvider{text: "Explicação remota."},
	)
	text, err := provider.Explain(context.Background(), ExplanationInput{})
	if err != nil || text != "Explicação remota." {
		t.Fatalf("expected fallback explanation, got %q, %v", text, err)
	}
}
