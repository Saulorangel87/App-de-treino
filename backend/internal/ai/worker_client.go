package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// WorkerClient calls the protected Cadencia route on the Cloudflare Worker.
// It sends the same narrow, already-validated facts used by Ollama.
type WorkerClient struct {
	endpoint      string
	token         string
	client        *http.Client
	maxConcurrent chan struct{}
}

func NewWorkerClient(baseURL, token string, timeout time.Duration, maxConcurrent int) (*WorkerClient, error) {
	parsed, err := url.Parse(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, fmt.Errorf("invalid AI Worker URL")
	}
	if strings.TrimSpace(token) == "" || timeout < time.Second || maxConcurrent < 1 || maxConcurrent > 2 {
		return nil, fmt.Errorf("invalid AI Worker limits")
	}
	return &WorkerClient{
		endpoint:      parsed.String() + "/cadencia/explanation",
		token:         strings.TrimSpace(token),
		client:        &http.Client{Timeout: timeout},
		maxConcurrent: make(chan struct{}, maxConcurrent),
	}, nil
}

type workerExplanationRequest struct {
	WorkoutName   string   `json:"workout_name"`
	Objective     string   `json:"objective"`
	Duration      int      `json:"duration_minutes"`
	TargetRPE     float64  `json:"target_rpe"`
	Rules         []string `json:"rules,omitempty"`
	EvidenceScope string   `json:"evidence_scope,omitempty"`
}

type workerExplanationResponse struct {
	Explanation string `json:"explanation"`
}

func (c *WorkerClient) Explain(ctx context.Context, input ExplanationInput) (string, error) {
	started := time.Now()
	var resultErr error
	defer func() { logProviderRequest("worker", started, resultErr) }()

	if err := validateInput(input); err != nil {
		resultErr = err
		return "", err
	}
	select {
	case c.maxConcurrent <- struct{}{}:
		defer func() { <-c.maxConcurrent }()
	case <-ctx.Done():
		resultErr = ctx.Err()
		return "", ctx.Err()
	}

	payload, err := json.Marshal(workerExplanationRequest{
		WorkoutName:   input.WorkoutName,
		Objective:     input.Objective,
		Duration:      input.DurationMinutes,
		TargetRPE:     input.TargetRPE,
		Rules:         input.Rules,
		EvidenceScope: input.EvidenceScope,
	})
	if err != nil {
		resultErr = err
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		resultErr = err
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+c.token)
	response, err := c.client.Do(request)
	if err != nil {
		resultErr = fmt.Errorf("%w: Worker request failed: %v", ErrUnavailable, err)
		return "", resultErr
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		resultErr = fmt.Errorf("%w: Worker returned %d", ErrUnavailable, response.StatusCode)
		return "", resultErr
	}

	var result workerExplanationResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 16*1024)).Decode(&result); err != nil {
		resultErr = fmt.Errorf("%w: %v", ErrInvalidResponse, err)
		return "", resultErr
	}
	text := strings.TrimSpace(result.Explanation)
	if text == "" || len([]rune(text)) > 1200 {
		resultErr = ErrInvalidResponse
		return "", ErrInvalidResponse
	}
	return text, nil
}

// FallbackProvider keeps deterministic rules as the final fallback while
// allowing a remote provider to be used only after the primary has failed.
type FallbackProvider struct {
	primary  Provider
	fallback Provider
}

func NewFallbackProvider(primary, fallback Provider) *FallbackProvider {
	return &FallbackProvider{primary: primary, fallback: fallback}
}

func (p *FallbackProvider) Explain(ctx context.Context, input ExplanationInput) (string, error) {
	if p == nil || p.primary == nil {
		if p != nil && p.fallback != nil {
			return p.fallback.Explain(ctx, input)
		}
		return "", ErrDisabled
	}
	text, primaryErr := p.primary.Explain(ctx, input)
	if primaryErr == nil {
		return text, nil
	}
	if p.fallback == nil {
		return "", primaryErr
	}
	text, fallbackErr := p.fallback.Explain(ctx, input)
	if fallbackErr == nil {
		return text, nil
	}
	return "", fmt.Errorf("%w: primary=%v; fallback=%v", ErrUnavailable, primaryErr, fallbackErr)
}
