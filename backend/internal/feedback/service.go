package feedback

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalidInput = errors.New("invalid user feedback")

const (
	CategoryExperience = "experience"
	CategoryBug        = "bug"
	CategorySuggestion = "suggestion"
)

// Entry is a product feedback report submitted by an authenticated athlete.
// UserID is kept for storage ownership but is intentionally not exposed in the
// API response used by the athlete-facing form.
type Entry struct {
	ID        string    `json:"id"`
	UserID    string    `json:"-"`
	Category  string    `json:"category"`
	Rating    int       `json:"rating"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateInput struct {
	Category string `json:"category"`
	Rating   int    `json:"rating"`
	Message  string `json:"message"`
}

// DigestEntry is the small, owner-only projection used by the weekly digest.
// It is never returned by the athlete-facing feedback endpoint.
type DigestEntry struct {
	ID          string
	DisplayName string
	Category    string
	Rating      int
	Message     string
	CreatedAt   time.Time
}

func CategoryLabel(value string) string {
	switch value {
	case CategoryExperience:
		return "Experiência"
	case CategoryBug:
		return "Problema"
	case CategorySuggestion:
		return "Sugestão"
	default:
		return "Feedback"
	}
}

type Store interface {
	CreateUserFeedback(context.Context, string, CreateInput) (Entry, error)
}

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (Entry, error) {
	input.Category = strings.ToLower(strings.TrimSpace(input.Category))
	input.Message = strings.TrimSpace(input.Message)
	if userID == "" || !validCategory(input.Category) || input.Rating < 1 || input.Rating > 5 || utf8.RuneCountInString(input.Message) < 10 || utf8.RuneCountInString(input.Message) > 2000 {
		return Entry{}, ErrInvalidInput
	}
	return s.store.CreateUserFeedback(ctx, userID, input)
}

func validCategory(value string) bool {
	return value == CategoryExperience || value == CategoryBug || value == CategorySuggestion
}
