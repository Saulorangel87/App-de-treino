package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrDisabled        = errors.New("ai explanations are disabled")
	ErrInvalidInput    = errors.New("invalid explanation input")
	ErrUnavailable     = errors.New("ai provider is unavailable")
	ErrInvalidResponse = errors.New("ai provider returned an invalid response")
)

// ExplanationInput is deliberately narrower than a full athlete profile or
// plan. The language model receives only facts already validated by rules-v1.
type ExplanationInput struct {
	WorkoutName     string
	Objective       string
	DurationMinutes int
	TargetRPE       float64
	Rules           []string
	EvidenceScope   string
}

type Provider interface {
	Explain(context.Context, ExplanationInput) (string, error)
}

type Service struct {
	provider Provider
}

func NewService(provider Provider) *Service {
	return &Service{provider: provider}
}

func (s *Service) Enabled() bool {
	return s != nil && s.provider != nil
}

func (s *Service) Explain(ctx context.Context, input ExplanationInput) (string, error) {
	if !s.Enabled() {
		return "", ErrDisabled
	}
	return s.provider.Explain(ctx, input)
}

type OllamaClient struct {
	baseURL         string
	model           string
	client          *http.Client
	maxOutputTokens int
	concurrency     chan struct{}
}

func NewOllamaClient(baseURL, model string, timeout time.Duration, maxOutputTokens, maxConcurrent int) (*OllamaClient, error) {
	parsed, err := url.Parse(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
	if err != nil || parsed.Scheme != "http" && parsed.Scheme != "https" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid Ollama base URL")
	}
	if strings.TrimSpace(model) == "" || timeout < time.Second || maxOutputTokens < 32 || maxOutputTokens > 512 || maxConcurrent < 1 || maxConcurrent > 2 {
		return nil, fmt.Errorf("invalid Ollama limits")
	}
	return &OllamaClient{
		baseURL:         parsed.String(),
		model:           model,
		client:          &http.Client{Timeout: timeout},
		maxOutputTokens: maxOutputTokens,
		concurrency:     make(chan struct{}, maxConcurrent),
	}, nil
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  map[string]any  `json:"options"`
}

type ollamaChatResponse struct {
	Message ollamaMessage `json:"message"`
}

func (c *OllamaClient) Explain(ctx context.Context, input ExplanationInput) (string, error) {
	if err := validateInput(input); err != nil {
		return "", err
	}
	select {
	case c.concurrency <- struct{}{}:
		defer func() { <-c.concurrency }()
	case <-ctx.Done():
		return "", ctx.Err()
	}

	payload, err := json.Marshal(ollamaChatRequest{
		Model: c.model,
		Messages: []ollamaMessage{
			{Role: "system", Content: "Você é o assistente explicativo do Cadência. Explique somente a sessão apresentada com base nos fatos fornecidos. Nunca altere duração, RPE ou estrutura, não invente estudos ou referências, não faça diagnóstico e não prescreva tratamento. Responda em português do Brasil, em duas ou três frases curtas, com linguagem amigável. Se houver um aviso de segurança nas regras, destaque-o sem minimizar o aviso."},
			{Role: "user", Content: explanationPrompt(input)},
		},
		Stream:  false,
		Options: map[string]any{"num_predict": c.maxOutputTokens, "temperature": 0.2},
	})
	if err != nil {
		return "", err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
		return "", fmt.Errorf("%w: Ollama returned %d: %s", ErrUnavailable, response.StatusCode, strings.TrimSpace(string(body)))
	}
	var result ollamaChatResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 16*1024)).Decode(&result); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}
	text := strings.TrimSpace(result.Message.Content)
	if text == "" || len([]rune(text)) > 1200 {
		return "", ErrInvalidResponse
	}
	return text, nil
}

func validateInput(input ExplanationInput) error {
	if strings.TrimSpace(input.WorkoutName) == "" || strings.TrimSpace(input.Objective) == "" || input.DurationMinutes < 1 || input.DurationMinutes > 480 || input.TargetRPE < 0 || input.TargetRPE > 10 {
		return ErrInvalidInput
	}
	if len(input.Rules) > 12 || len([]rune(input.EvidenceScope)) > 500 {
		return ErrInvalidInput
	}
	for _, rule := range input.Rules {
		if len([]rune(rule)) > 240 {
			return ErrInvalidInput
		}
	}
	return nil
}

func explanationPrompt(input ExplanationInput) string {
	rules := make([]string, 0, len(input.Rules))
	for _, rule := range input.Rules {
		rules = append(rules, "- "+rule)
	}
	return fmt.Sprintf("Explique a escolha desta sessão usando somente estes fatos validados:\nTreino: %s\nObjetivo: %s\nDuração: %d minutos\nRPE-alvo: %.1f\nRegras:\n%s\nEscopo das evidências: %s", input.WorkoutName, input.Objective, input.DurationMinutes, input.TargetRPE, strings.Join(rules, "\n"), input.EvidenceScope)
}
