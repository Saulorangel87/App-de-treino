package ai

import (
	"strings"
	"testing"
	"time"
)

func TestOnlyProvidersThatSayItAnswerInEnglish(t *testing.T) {
	ollama, err := NewOllamaClient("http://127.0.0.1:11434", "model", 5*time.Second, 128, 1)
	if err != nil {
		t.Fatalf("ollama client: %v", err)
	}
	worker, err := NewWorkerClient("https://worker.example", "token", 5*time.Second, 1)
	if err != nil {
		t.Fatalf("worker client: %v", err)
	}
	if !NewService(ollama).Supports("en") || !NewService(ollama).Supports("pt") {
		t.Fatal("the local model answers in both languages")
	}
	if NewService(worker).Supports("en") || !NewService(worker).Supports("pt") {
		t.Fatal("the Worker prompt is Portuguese only until it learns the language field")
	}
	if NewService(nil).Supports("en") {
		t.Fatal("without a provider there is nothing to answer in English")
	}
}

func TestPromptFollowsTheLanguage(t *testing.T) {
	input := ExplanationInput{Language: "en", WorkoutName: "Base ride", Objective: "Develop general fitness", DurationMinutes: 60, TargetRPE: 4, Rules: []string{"Rule"}}
	if prompt := explanationPrompt(input); !strings.HasPrefix(prompt, "Explain why this session was chosen") || !strings.Contains(prompt, "Target RPE: 4.0") {
		t.Fatalf("English prompt: %q", prompt)
	}
	if !strings.Contains(systemPromptFor("en"), "Answer in English") || !strings.Contains(systemPromptFor("pt"), "português do Brasil") {
		t.Fatal("system prompt must follow the language")
	}
	input.Language = ""
	if !strings.HasPrefix(explanationPrompt(input), "Explique a escolha") {
		t.Fatal("Portuguese stays the default")
	}
}
