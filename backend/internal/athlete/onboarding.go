package athlete

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrInvalidOnboarding = errors.New("invalid onboarding data")

type Limitation struct {
	Kind                             string `json:"kind"`
	Description                      string `json:"description"`
	IsActive                         bool   `json:"is_active"`
	ProfessionalClearanceRecommended bool   `json:"professional_clearance_recommended"`
}

type Goal struct {
	GoalType   string  `json:"goal_type"`
	Priority   int     `json:"priority"`
	TargetDate *string `json:"target_date"`
	Details    string  `json:"details"`
}

type Availability struct {
	Weekday          int     `json:"weekday"`
	AvailableMinutes int     `json:"available_minutes"`
	PreferredTime    *string `json:"preferred_time"`
	Location         *string `json:"location"`
}

type Onboarding struct {
	Limitations    []Limitation   `json:"limitations"`
	Goals          []Goal         `json:"goals"`
	Availability   []Availability `json:"availability"`
	CyclingContext CyclingContext `json:"cycling_context"`
}

type CyclingContext struct {
	WeeklyHours            float64 `json:"weekly_hours"`
	LongestRideMinutes     int     `json:"longest_ride_minutes"`
	WeeklyRides            int     `json:"weekly_rides"`
	RecentWeeklyDistanceKM float64 `json:"recent_weekly_distance_km"`
	BikeType               string  `json:"bike_type"`
	Terrain                string  `json:"terrain"`
	UsesHeartRate          bool    `json:"uses_heart_rate"`
	UsesPower              bool    `json:"uses_power"`
	FTP                    *int    `json:"ftp,omitempty"`
	EventGoal              bool    `json:"event_goal"`
	EventDistanceKM        *int    `json:"event_distance_km,omitempty"`
	EventDate              *string `json:"event_date,omitempty"`
}

type OnboardingStore interface {
	OnboardingByUserID(context.Context, string) (Onboarding, error)
	ReplaceLimitations(context.Context, string, []Limitation) ([]Limitation, error)
	ReplaceGoals(context.Context, string, []Goal) ([]Goal, error)
	ReplaceAvailability(context.Context, string, []Availability) ([]Availability, error)
	SaveCyclingContext(context.Context, string, CyclingContext) (CyclingContext, error)
}

func (s *OnboardingService) SaveCyclingContext(ctx context.Context, userID string, value CyclingContext) (CyclingContext, error) {
	if value.WeeklyHours < 0 || value.WeeklyHours > 80 || value.LongestRideMinutes < 0 || value.LongestRideMinutes > 1440 || value.WeeklyRides < 0 || value.WeeklyRides > 21 || value.RecentWeeklyDistanceKM < 0 || value.RecentWeeklyDistanceKM > 2000 || (value.FTP != nil && (*value.FTP < 50 || *value.FTP > 600)) || (value.EventDistanceKM != nil && (*value.EventDistanceKM < 1 || *value.EventDistanceKM > 2000)) {
		return CyclingContext{}, ErrInvalidOnboarding
	}
	if value.FTP != nil && !value.UsesPower {
		return CyclingContext{}, ErrInvalidOnboarding
	}
	if value.EventGoal {
		if value.EventDistanceKM == nil || value.EventDate == nil {
			return CyclingContext{}, ErrInvalidOnboarding
		}
		if _, err := time.Parse("2006-01-02", *value.EventDate); err != nil {
			return CyclingContext{}, ErrInvalidOnboarding
		}
	} else {
		value.EventDistanceKM, value.EventDate = nil, nil
	}
	return s.store.SaveCyclingContext(ctx, userID, value)
}

type OnboardingService struct{ store OnboardingStore }

func NewOnboardingService(store OnboardingStore) *OnboardingService {
	return &OnboardingService{store: store}
}

func (s *OnboardingService) Get(ctx context.Context, userID string) (Onboarding, error) {
	return s.store.OnboardingByUserID(ctx, userID)
}

func (s *OnboardingService) SaveLimitations(ctx context.Context, userID string, limitations []Limitation) ([]Limitation, error) {
	if len(limitations) > 10 {
		return nil, ErrInvalidOnboarding
	}
	kinds := map[string]bool{"pain": true, "injury": true, "medical_condition": true, "mobility": true, "other": true}
	for index := range limitations {
		limitations[index].Kind = strings.TrimSpace(limitations[index].Kind)
		limitations[index].Description = strings.TrimSpace(limitations[index].Description)
		if !kinds[limitations[index].Kind] || len(limitations[index].Description) < 3 || len(limitations[index].Description) > 500 {
			return nil, ErrInvalidOnboarding
		}
		limitations[index].IsActive = true
	}
	return s.store.ReplaceLimitations(ctx, userID, limitations)
}

func (s *OnboardingService) SaveGoals(ctx context.Context, userID string, goals []Goal) ([]Goal, error) {
	if len(goals) < 1 || len(goals) > 2 {
		return nil, ErrInvalidOnboarding
	}
	types := map[string]bool{"health": true, "fitness": true, "endurance": true, "performance": true, "event": true, "weight_management": true}
	priorities := map[int]bool{}
	for index := range goals {
		goals[index].GoalType = strings.TrimSpace(goals[index].GoalType)
		goals[index].Details = strings.TrimSpace(goals[index].Details)
		if !types[goals[index].GoalType] || goals[index].Priority < 1 || goals[index].Priority > 2 || priorities[goals[index].Priority] || len(goals[index].Details) > 500 {
			return nil, ErrInvalidOnboarding
		}
		priorities[goals[index].Priority] = true
		if goals[index].TargetDate != nil {
			date, err := time.Parse("2006-01-02", *goals[index].TargetDate)
			if err != nil || date.Before(time.Now().AddDate(0, 0, -1)) {
				return nil, ErrInvalidOnboarding
			}
		}
	}
	if !priorities[1] {
		return nil, ErrInvalidOnboarding
	}
	return s.store.ReplaceGoals(ctx, userID, goals)
}

func (s *OnboardingService) SaveAvailability(ctx context.Context, userID string, availability []Availability) ([]Availability, error) {
	if len(availability) != 7 {
		return nil, ErrInvalidOnboarding
	}
	weekdays := map[int]bool{}
	totalMinutes := 0
	for index := range availability {
		item := &availability[index]
		if item.Weekday < 0 || item.Weekday > 6 || weekdays[item.Weekday] || item.AvailableMinutes < 0 || item.AvailableMinutes > 480 {
			return nil, ErrInvalidOnboarding
		}
		weekdays[item.Weekday] = true
		totalMinutes += item.AvailableMinutes
		if item.PreferredTime != nil {
			if _, err := time.Parse("15:04", *item.PreferredTime); err != nil {
				return nil, ErrInvalidOnboarding
			}
		}
		if item.Location != nil {
			location := strings.TrimSpace(*item.Location)
			if len(location) > 80 {
				return nil, ErrInvalidOnboarding
			}
			if location == "" {
				item.Location = nil
			} else {
				item.Location = &location
			}
		}
	}
	if totalMinutes == 0 {
		return nil, ErrInvalidOnboarding
	}
	return s.store.ReplaceAvailability(ctx, userID, availability)
}
