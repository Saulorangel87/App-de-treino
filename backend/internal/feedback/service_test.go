package feedback

import (
	"context"
	"testing"
)

type fakeStore struct {
	userID string
	input  CreateInput
}

func (s *fakeStore) CreateUserFeedback(_ context.Context, userID string, input CreateInput) (Entry, error) {
	s.userID, s.input = userID, input
	return Entry{ID: "feedback-1", UserID: userID, Category: input.Category, Rating: input.Rating, Message: input.Message}, nil
}

func TestCreateNormalizesAndDelegatesFeedback(t *testing.T) {
	store := &fakeStore{}
	entry, err := NewService(store).Create(context.Background(), "user-1", CreateInput{
		Category: " Suggestion ",
		Rating:   4,
		Message:  " A experiência foi clara e útil. ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.userID != "user-1" || store.input.Category != CategorySuggestion || store.input.Message != "A experiência foi clara e útil." || entry.ID != "feedback-1" {
		t.Fatalf("unexpected feedback: %#v, stored %#v", entry, store)
	}
}

func TestCreateRejectsInvalidFeedback(t *testing.T) {
	store := &fakeStore{}
	_, err := NewService(store).Create(context.Background(), "user-1", CreateInput{Category: "other", Rating: 6, Message: "curto"})
	if err != ErrInvalidInput {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if store.userID != "" {
		t.Fatal("store must not receive invalid feedback")
	}
}
