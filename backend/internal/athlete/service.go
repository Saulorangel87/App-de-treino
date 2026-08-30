package athlete

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidProfile = errors.New("invalid profile")
	ErrProfileMissing = errors.New("profile not found")
)

type Profile struct {
	UserID          string   `json:"user_id"`
	BirthDate       *string  `json:"birth_date"`
	Sex             *string  `json:"sex"`
	HeightCM        *float64 `json:"height_cm"`
	WeightKG        *float64 `json:"weight_kg"`
	Sport           string   `json:"sport"`
	ExperienceLevel string   `json:"experience_level"`
	ActivityLevel   *string  `json:"activity_level"`
}

type Store interface {
	UpsertProfile(context.Context, Profile) (Profile, error)
	ProfileByUserID(context.Context, string) (Profile, error)
}

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) Save(ctx context.Context, profile Profile) (Profile, error) {
	if !validProfile(profile) {
		return Profile{}, ErrInvalidProfile
	}
	profile.Sport = "cycling"
	return s.store.UpsertProfile(ctx, profile)
}

func (s *Service) Get(ctx context.Context, userID string) (Profile, error) {
	return s.store.ProfileByUserID(ctx, userID)
}

func validProfile(profile Profile) bool {
	levels := map[string]bool{"beginner": true, "intermediate": true, "advanced": true}
	sexes := map[string]bool{"female": true, "male": true, "other": true, "prefer_not_to_say": true}
	if !levels[profile.ExperienceLevel] {
		return false
	}
	if profile.Sex != nil && !sexes[*profile.Sex] {
		return false
	}
	if profile.HeightCM != nil && (*profile.HeightCM < 100 || *profile.HeightCM > 250) {
		return false
	}
	if profile.WeightKG != nil && (*profile.WeightKG < 30 || *profile.WeightKG > 350) {
		return false
	}
	if profile.BirthDate != nil {
		birthDate, err := time.Parse("2006-01-02", *profile.BirthDate)
		if err != nil || birthDate.After(time.Now()) || birthDate.Before(time.Now().AddDate(-120, 0, 0)) {
			return false
		}
	}
	return true
}
