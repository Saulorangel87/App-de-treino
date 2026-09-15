package athlete

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrInvalidOnboarding = errors.New("invalid onboarding data")

type Limitation struct {
	Kind                             string   `json:"kind"`
	Description                      string   `json:"description"`
	Location                         string   `json:"location,omitempty"`
	Intensity                        *int     `json:"intensity,omitempty"`
	AggravatingMovement              string   `json:"aggravating_movement,omitempty"`
	StartedOn                        *string  `json:"started_on,omitempty"`
	SymptomsDuringAfter              []string `json:"symptoms_during_after,omitempty"`
	MedicalRestriction               bool     `json:"medical_restriction"`
	RecentSurgery                    bool     `json:"recent_surgery"`
	ExerciseProhibited               bool     `json:"exercise_prohibited"`
	ConditionAffectingExercise       bool     `json:"condition_affecting_exercise"`
	IsActive                         bool     `json:"is_active"`
	ProfessionalClearanceRecommended bool     `json:"professional_clearance_recommended"`
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
	WeeklyHours            float64  `json:"weekly_hours"`
	PracticeDurationMonths int      `json:"practice_duration_months"`
	AverageRideMinutes     int      `json:"average_ride_minutes"`
	LongestRideMinutes     int      `json:"longest_ride_minutes"`
	WeeklyRides            int      `json:"weekly_rides"`
	RecentWeeklyDistanceKM float64  `json:"recent_weekly_distance_km"`
	RecentTrainingWeeks    int      `json:"recent_training_weeks"`
	TrainingStatus         string   `json:"training_status"`
	RecentBestDistanceKM   float64  `json:"recent_best_distance_km"`
	PreferredSessionTypes  []string `json:"preferred_session_types"`
	Discipline             string   `json:"discipline"`
	BikeType               string   `json:"bike_type"`
	Terrain                string   `json:"terrain"`
	UsesHeartRate          bool     `json:"uses_heart_rate"`
	UsesPower              bool     `json:"uses_power"`
	UsesGPS                bool     `json:"uses_gps"`
	UsesSportsWatch        bool     `json:"uses_sports_watch"`
	UsesSmartTrainer       bool     `json:"uses_smart_trainer"`
	FTP                    *int     `json:"ftp,omitempty"`
	FTPTestDate            *string  `json:"ftp_test_date,omitempty"`
	FTPProtocol            string   `json:"ftp_protocol,omitempty"`
	AveragePowerWatts      *int     `json:"average_power_watts,omitempty"`
	EventGoal              bool     `json:"event_goal"`
	EventDistanceKM        *int     `json:"event_distance_km,omitempty"`
	EventDate              *string  `json:"event_date,omitempty"`
}

type OnboardingStore interface {
	OnboardingByUserID(context.Context, string) (Onboarding, error)
	ReplaceLimitations(context.Context, string, []Limitation) ([]Limitation, error)
	ReplaceGoals(context.Context, string, []Goal) ([]Goal, error)
	ReplaceAvailability(context.Context, string, []Availability) ([]Availability, error)
	SaveCyclingContext(context.Context, string, CyclingContext) (CyclingContext, error)
}

func (s *OnboardingService) SaveCyclingContext(ctx context.Context, userID string, value CyclingContext) (CyclingContext, error) {
	if value.PreferredSessionTypes == nil {
		value.PreferredSessionTypes = []string{}
	}
	value.Discipline = strings.TrimSpace(value.Discipline)
	value.TrainingStatus = strings.TrimSpace(value.TrainingStatus)
	if value.TrainingStatus == "" {
		value.TrainingStatus = "not_informed"
	}
	if value.WeeklyHours < 0 || value.WeeklyHours > 80 || value.PracticeDurationMonths < 0 || value.PracticeDurationMonths > 1200 || value.AverageRideMinutes < 0 || value.AverageRideMinutes > 1440 || value.LongestRideMinutes < 0 || value.LongestRideMinutes > 1440 || value.WeeklyRides < 0 || value.WeeklyRides > 21 || value.RecentWeeklyDistanceKM < 0 || value.RecentWeeklyDistanceKM > 2000 || value.RecentTrainingWeeks < 0 || value.RecentTrainingWeeks > 52 || value.RecentBestDistanceKM < 0 || value.RecentBestDistanceKM > 2000 || len(value.PreferredSessionTypes) > 9 || (value.FTP != nil && (*value.FTP < 50 || *value.FTP > 600)) || (value.AveragePowerWatts != nil && (*value.AveragePowerWatts < 0 || *value.AveragePowerWatts > 2000)) || (value.EventDistanceKM != nil && (*value.EventDistanceKM < 1 || *value.EventDistanceKM > 2000)) {
		return CyclingContext{}, ErrInvalidOnboarding
	}
	allowedDisciplines := map[string]bool{"": true, "general": true, "road": true, "mtb_xco": true, "mtb_xcm": true, "gravel": true, "indoor": true}
	if !allowedDisciplines[value.Discipline] {
		return CyclingContext{}, ErrInvalidOnboarding
	}
	allowedTrainingStatuses := map[string]bool{"not_informed": true, "regular": true, "returning_after_break": true}
	if !allowedTrainingStatuses[value.TrainingStatus] {
		return CyclingContext{}, ErrInvalidOnboarding
	}
	allowedPreferences := map[string]bool{"base": true, "cadence": true, "hills": true, "intervals": true, "threshold": true, "sweet_spot": true, "vo2max": true, "short_intervals": true, "recovery": true}
	seenPreferences := map[string]bool{}
	for index := range value.PreferredSessionTypes {
		preference := strings.TrimSpace(value.PreferredSessionTypes[index])
		if !allowedPreferences[preference] || seenPreferences[preference] {
			return CyclingContext{}, ErrInvalidOnboarding
		}
		value.PreferredSessionTypes[index] = preference
		seenPreferences[preference] = true
	}
	if !value.UsesPower {
		value.FTP, value.FTPTestDate, value.FTPProtocol, value.AveragePowerWatts = nil, nil, "", nil
	} else {
		if value.FTPTestDate != nil {
			date, err := time.Parse("2006-01-02", strings.TrimSpace(*value.FTPTestDate))
			now := s.now()
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			if err != nil || date.After(today) {
				return CyclingContext{}, ErrInvalidOnboarding
			}
			formatted := date.Format("2006-01-02")
			value.FTPTestDate = &formatted
		}
		if value.FTPProtocol != "" && value.FTPProtocol != "20_minute" && value.FTPProtocol != "ramp" && value.FTPProtocol != "other" {
			return CyclingContext{}, ErrInvalidOnboarding
		}
	}
	if value.EventGoal {
		if value.EventDistanceKM == nil || value.EventDate == nil {
			return CyclingContext{}, ErrInvalidOnboarding
		}
		eventDate, err := time.ParseInLocation("2006-01-02", *value.EventDate, time.Local)
		today := s.now()
		todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
		if err != nil || eventDate.Before(todayDate) {
			return CyclingContext{}, ErrInvalidOnboarding
		}
	} else {
		value.EventDistanceKM, value.EventDate = nil, nil
	}
	return s.store.SaveCyclingContext(ctx, userID, value)
}

type OnboardingService struct {
	store OnboardingStore
	now   func() time.Time
}

func NewOnboardingService(store OnboardingStore) *OnboardingService {
	return &OnboardingService{store: store, now: time.Now}
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
		limitations[index].Location = strings.TrimSpace(limitations[index].Location)
		limitations[index].AggravatingMovement = strings.TrimSpace(limitations[index].AggravatingMovement)
		if limitations[index].SymptomsDuringAfter == nil {
			limitations[index].SymptomsDuringAfter = []string{}
		}
		if !kinds[limitations[index].Kind] || len(limitations[index].Description) < 3 || len(limitations[index].Description) > 500 || len(limitations[index].Location) > 120 || len(limitations[index].AggravatingMovement) > 200 || len(limitations[index].SymptomsDuringAfter) > 5 {
			return nil, ErrInvalidOnboarding
		}
		allowedSymptoms := map[string]bool{"dizziness": true, "unusual_shortness_of_breath": true, "malaise": true, "extreme_fatigue": true, "other": true}
		seenSymptoms := map[string]bool{}
		for symptomIndex, symptom := range limitations[index].SymptomsDuringAfter {
			symptom = strings.TrimSpace(symptom)
			if !allowedSymptoms[symptom] || seenSymptoms[symptom] {
				return nil, ErrInvalidOnboarding
			}
			limitations[index].SymptomsDuringAfter[symptomIndex] = symptom
			seenSymptoms[symptom] = true
		}
		if limitations[index].Intensity != nil && (*limitations[index].Intensity < 1 || *limitations[index].Intensity > 10) {
			return nil, ErrInvalidOnboarding
		}
		if limitations[index].StartedOn != nil {
			startedOn, err := time.Parse("2006-01-02", strings.TrimSpace(*limitations[index].StartedOn))
			today := s.now()
			todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
			if err != nil || startedOn.After(todayDate) {
				return nil, ErrInvalidOnboarding
			}
			startedOnValue := startedOn.Format("2006-01-02")
			limitations[index].StartedOn = &startedOnValue
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
			if err != nil || date.Before(s.now().AddDate(0, 0, -1)) {
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
