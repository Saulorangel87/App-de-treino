package email

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var ErrNotConfigured = errors.New("email sender is not configured")

type Message struct {
	To      string
	Subject string
	HTML    string
	Text    string
}

type Sender interface {
	Send(context.Context, Message) error
}

type DisabledSender struct{}

func (DisabledSender) Send(context.Context, Message) error { return ErrNotConfigured }

// DevelopmentSender intentionally does not deliver email. It allows local-only
// verification through the URL returned by the development API response.
type DevelopmentSender struct{}

func (DevelopmentSender) Send(context.Context, Message) error { return nil }

type ResendSender struct {
	apiKey string
	from   string
	client *http.Client
}

func NewResendSender(apiKey, from string) *ResendSender {
	return &ResendSender{apiKey: apiKey, from: from, client: &http.Client{Timeout: 10 * time.Second}}
}

func (s *ResendSender) Send(ctx context.Context, message Message) error {
	payload, err := json.Marshal(map[string]any{
		"from": s.from, "to": []string{message.To}, "subject": message.Subject,
		"html": message.HTML, "text": message.Text,
	})
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+s.apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "cadencia-api/0.5")
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	return fmt.Errorf("resend returned %d: %s", response.StatusCode, string(body))
}
