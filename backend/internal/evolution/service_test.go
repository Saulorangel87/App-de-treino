package evolution

import (
	"context"
	"testing"
)

type fakeStore struct {
	userID string
	result Summary
}

func (s *fakeStore) EvolutionSummaryByUserID(_ context.Context, userID string) (Summary, error) {
	s.userID = userID
	return s.result, nil
}

func TestSummaryDelegatesToTheAuthenticatedUserStore(t *testing.T) {
	store := &fakeStore{result: Summary{CompletedSessions: 3}}
	result, err := NewService(store).Summary(context.Background(), "user-1")
	if err != nil || store.userID != "user-1" || result.CompletedSessions != 3 {
		t.Fatalf("unexpected summary result: %#v, user %q, error %v", result, store.userID, err)
	}
}
