package athlete

import (
	"context"
	"testing"
)

type profileStore struct{ saved Profile }

func (s *profileStore) UpsertProfile(_ context.Context, profile Profile) (Profile, error) {
	s.saved = profile
	return profile, nil
}
func (s *profileStore) ProfileByUserID(_ context.Context, _ string) (Profile, error) {
	return s.saved, nil
}

func TestSaveProfileForcesCyclingAndValidatesMeasurements(t *testing.T) {
	store := &profileStore{}
	service := NewService(store)
	height := 178.0
	weight := 76.0
	profile, err := service.Save(context.Background(), Profile{UserID: "user-1", HeightCM: &height, WeightKG: &weight, ExperienceLevel: "intermediate"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Sport != "cycling" {
		t.Fatalf("expected cycling, got %s", profile.Sport)
	}
}

func TestSaveRejectsInvalidHeight(t *testing.T) {
	height := 40.0
	_, err := NewService(&profileStore{}).Save(context.Background(), Profile{HeightCM: &height, ExperienceLevel: "beginner"})
	if err != ErrInvalidProfile {
		t.Fatalf("expected ErrInvalidProfile, got %v", err)
	}
}
