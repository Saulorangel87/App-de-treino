package planning

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"time"
)

var (
	ErrIncompleteOnboarding = errors.New("incomplete onboarding")
	ErrInvalidPlanID        = errors.New("invalid plan id")
	ErrPlanMissing          = errors.New("plan not found")
	ErrInvalidWorkoutID     = errors.New("invalid workout id")
	ErrWorkoutMissing       = errors.New("workout not found")
	ErrInvalidTransition    = errors.New("invalid workout transition")
	ErrInvalidFeedback      = errors.New("invalid workout feedback")
)

var planIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

type LimitationContext struct {
	Kind                             string
	ProfessionalClearanceRecommended bool
}

type AvailabilitySlot struct {
	Weekday          int
	AvailableMinutes int
	Location         *string
}

type CyclingContext struct {
	BikeType  string `json:"bike_type"`
	Terrain   string `json:"terrain"`
	UsesPower bool   `json:"uses_power"`
	FTP       *int   `json:"ftp,omitempty"`
	EventGoal bool   `json:"event_goal"`
}

type Context struct {
	ProfileID        string
	ExperienceLevel  string
	PrimaryGoal      string
	Limitations      []LimitationContext
	Availability     []AvailabilitySlot
	Cycling          CyclingContext
	BaselineEligible bool
	RotationIndex    int
}

type Workout struct {
	ID              string          `json:"id,omitempty"`
	ScheduledOn     string          `json:"scheduled_on"`
	Name            string          `json:"name"`
	Objective       string          `json:"objective"`
	DurationMinutes int             `json:"duration_minutes"`
	TargetRPE       float64         `json:"target_rpe"`
	Structure       map[string]any  `json:"structure"`
	Explanation     map[string]any  `json:"explanation"`
	Status          string          `json:"status"`
	Session         *WorkoutSession `json:"session,omitempty"`
}

type WorkoutSession struct {
	ID              string     `json:"id"`
	Status          string     `json:"status"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CancelledAt     *time.Time `json:"cancelled_at,omitempty"`
	DurationMinutes *int       `json:"duration_minutes,omitempty"`
	ActualRPE       *float64   `json:"actual_rpe,omitempty"`
	Feedback        *Feedback  `json:"feedback,omitempty"`
}

type Feedback struct {
	Difficulty   string `json:"difficulty"`
	PainReported bool   `json:"pain_reported"`
	FatigueAfter int    `json:"fatigue_after"`
	Notes        string `json:"notes,omitempty"`
}

type CompletionInput struct {
	ActualRPE    float64
	Difficulty   string
	PainReported bool
	FatigueAfter int
	Notes        string
}

type Activity struct {
	ID              string     `json:"id"`
	WorkoutID       string     `json:"workout_id"`
	Name            string     `json:"name"`
	Objective       string     `json:"objective"`
	ScheduledOn     string     `json:"scheduled_on"`
	Status          string     `json:"status"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CancelledAt     *time.Time `json:"cancelled_at,omitempty"`
	DurationMinutes *int       `json:"duration_minutes,omitempty"`
	ActualRPE       *float64   `json:"actual_rpe,omitempty"`
	Feedback        *Feedback  `json:"feedback,omitempty"`
}

type ScientificSource struct {
	SourceKey     string `json:"source_key"`
	Title         string `json:"title"`
	Authors       string `json:"authors"`
	PublishedYear int    `json:"published_year"`
	URL           string `json:"url"`
	TrainingFocus string `json:"training_focus"`
	EvidenceLevel string `json:"evidence_level"`
	Summary       string `json:"summary"`
}

type Plan struct {
	ID                   string             `json:"id,omitempty"`
	StartsOn             string             `json:"starts_on"`
	EndsOn               string             `json:"ends_on"`
	Status               string             `json:"status"`
	PrescriptionSnapshot map[string]any     `json:"prescription_snapshot"`
	Workouts             []Workout          `json:"workouts"`
	Evidence             []ScientificSource `json:"evidence,omitempty"`
}

type Store interface {
	PlanningContextByUserID(context.Context, string) (Context, error)
	SaveDraftPlan(context.Context, string, Plan) (Plan, error)
	CurrentPlanByUserID(context.Context, string) (Plan, error)
	ActivatePlanByUserID(context.Context, string, string) error
	StartWorkoutByUserID(context.Context, string, string) error
	CompleteWorkoutByUserID(context.Context, string, string, CompletionInput) error
	CancelWorkoutByUserID(context.Context, string, string) error
	ActivitiesByUserID(context.Context, string) ([]Activity, error)
}

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service {
	return &Service{store: store, now: time.Now}
}

func (s *Service) Generate(ctx context.Context, userID string) (Plan, error) {
	input, err := s.store.PlanningContextByUserID(ctx, userID)
	if err != nil {
		return Plan{}, err
	}
	plan, err := buildPlan(input, s.now())
	if err != nil {
		return Plan{}, err
	}
	return s.store.SaveDraftPlan(ctx, input.ProfileID, plan)
}

func (s *Service) Current(ctx context.Context, userID string) (Plan, error) {
	return s.store.CurrentPlanByUserID(ctx, userID)
}

func (s *Service) Activate(ctx context.Context, userID, planID string) (Plan, error) {
	if !planIDPattern.MatchString(planID) {
		return Plan{}, ErrInvalidPlanID
	}
	if err := s.store.ActivatePlanByUserID(ctx, userID, planID); err != nil {
		return Plan{}, err
	}
	return s.store.CurrentPlanByUserID(ctx, userID)
}

func (s *Service) StartWorkout(ctx context.Context, userID, workoutID string) (Plan, error) {
	if !planIDPattern.MatchString(workoutID) {
		return Plan{}, ErrInvalidWorkoutID
	}
	if err := s.store.StartWorkoutByUserID(ctx, userID, workoutID); err != nil {
		return Plan{}, err
	}
	return s.store.CurrentPlanByUserID(ctx, userID)
}

func (s *Service) CompleteWorkout(ctx context.Context, userID, workoutID string, input CompletionInput) (Plan, error) {
	if !planIDPattern.MatchString(workoutID) {
		return Plan{}, ErrInvalidWorkoutID
	}
	if !validCompletion(input) {
		return Plan{}, ErrInvalidFeedback
	}
	if err := s.store.CompleteWorkoutByUserID(ctx, userID, workoutID, input); err != nil {
		return Plan{}, err
	}
	return s.store.CurrentPlanByUserID(ctx, userID)
}

func (s *Service) CancelWorkout(ctx context.Context, userID, workoutID string) (Plan, error) {
	if !planIDPattern.MatchString(workoutID) {
		return Plan{}, ErrInvalidWorkoutID
	}
	if err := s.store.CancelWorkoutByUserID(ctx, userID, workoutID); err != nil {
		return Plan{}, err
	}
	return s.store.CurrentPlanByUserID(ctx, userID)
}

func (s *Service) Activities(ctx context.Context, userID string) ([]Activity, error) {
	return s.store.ActivitiesByUserID(ctx, userID)
}

func validCompletion(input CompletionInput) bool {
	if input.ActualRPE < 1 || input.ActualRPE > 10 || input.FatigueAfter < 1 || input.FatigueAfter > 5 || len(input.Notes) > 1000 {
		return false
	}
	switch input.Difficulty {
	case "very_easy", "easy", "moderate", "hard", "very_hard":
		return true
	default:
		return false
	}
}

func buildPlan(input Context, now time.Time) (Plan, error) {
	if input.ProfileID == "" || input.PrimaryGoal == "" || len(input.Availability) == 0 {
		return Plan{}, ErrIncompleteOnboarding
	}
	maxSessions := map[string]int{"beginner": 3, "intermediate": 4, "advanced": 5}[input.ExperienceLevel]
	if maxSessions == 0 {
		return Plan{}, ErrIncompleteOnboarding
	}

	slots := append([]AvailabilitySlot(nil), input.Availability...)
	sort.Slice(slots, func(i, j int) bool { return slots[i].AvailableMinutes > slots[j].AvailableMinutes })
	if len(slots) > maxSessions {
		slots = slots[:maxSessions]
	}
	sort.Slice(slots, func(i, j int) bool { return slots[i].Weekday < slots[j].Weekday })

	start := nextMonday(now)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	restricted := len(input.Limitations) > 0
	for _, item := range input.Limitations {
		if item.ProfessionalClearanceRecommended {
			restricted = true
		}
	}
	workouts := make([]Workout, 0, len(slots)*4)
	multipliers := []float64{0.85, 0.95, 1.0, 0.75}
	for week := 0; week < 4; week++ {
		longIndex := longestSlot(slots)
		intensityIndex := intensitySlot(slots, longIndex)
		for index, slot := range slots {
			scheduledOn := start.AddDate(0, 0, week*7+weekdayOffset(slot.Weekday))
			if week == 0 && scheduledOn.Before(today) {
				continue
			}
			kind := "base"
			if index == longIndex {
				kind = "long"
			} else if index == intensityIndex && !restricted {
				kind = "quality"
			}
			workouts = append(workouts, makeWorkout(input, slot, kind, restricted, multipliers[week], scheduledOn))
		}
	}

	plan := Plan{
		StartsOn: start.Format("2006-01-02"),
		EndsOn:   start.AddDate(0, 0, 27).Format("2006-01-02"),
		Status:   "draft",
		PrescriptionSnapshot: map[string]any{
			"engine_version":    "rules-v1",
			"experience_level":  input.ExperienceLevel,
			"primary_goal":      input.PrimaryGoal,
			"restricted":        restricted,
			"sessions_per_week": len(slots),
			"cycling_context": map[string]any{
				"bike_type":  input.Cycling.BikeType,
				"terrain":    input.Cycling.Terrain,
				"uses_power": input.Cycling.UsesPower,
				"event_goal": input.Cycling.EventGoal,
			},
			"baseline_eligible": input.BaselineEligible,
			"rotation_index":    input.RotationIndex,
		},
		Workouts: workouts,
	}
	return plan, nil
}

func makeWorkout(input Context, slot AvailabilitySlot, kind string, restricted bool, multiplier float64, date time.Time) Workout {
	baseMinutes := map[string]int{"beginner": 45, "intermediate": 60, "advanced": 75}[input.ExperienceLevel]
	name := "Giro de base"
	targetRPE := 4.0
	mainBlock := "Ritmo confortável e contínuo"
	summary := explanationFor(kind, restricted)
	usesControlledIntervals := false
	rotationApplied := false
	if kind == "long" {
		baseMinutes = map[string]int{"beginner": 75, "intermediate": 90, "advanced": 120}[input.ExperienceLevel]
		name = "Endurance contínuo"
		targetRPE = 5.0
		mainBlock = "Volume aeróbico estável"
	}
	if kind == "quality" {
		name = "Tempo controlado"
		targetRPE = 6.0
		mainBlock = "3 blocos sustentados com recuperação leve"
		if input.ExperienceLevel == "advanced" && input.BaselineEligible && (input.PrimaryGoal == "performance" || input.PrimaryGoal == "event") && slot.AvailableMinutes >= 50 && multiplier >= 0.95 {
			name = "Intervalos controlados"
			targetRPE = 7.0
			mainBlock = "4 blocos de 4 min em esforço forte-controlado, com 3 min leves entre os blocos"
			summary = "A referência submáxima concluída sem sinais de alerta permite uma progressão intervalada controlada."
			usesControlledIntervals = true
		} else if input.ExperienceLevel == "intermediate" && input.Cycling.BikeType == "indoor" {
			name = "Cadência técnica"
			targetRPE = 5.0
			mainBlock = "6 blocos de cadência controlada com recuperação leve"
			summary = "A sessão usa o ambiente indoor informado para praticar cadência com esforço controlado."
		} else if input.Cycling.Terrain == "hilly" && input.ExperienceLevel != "beginner" {
			name = "Subidas controladas"
			if input.ExperienceLevel == "advanced" {
				targetRPE = 6.5
			}
			mainBlock = "4 blocos sustentados em subida, com recuperação leve"
			summary = "O terreno com subidas informado orienta um estímulo controlado e específico."
		} else if input.ExperienceLevel == "advanced" && input.Cycling.EventGoal {
			name = "Ritmo de prova controlado"
			targetRPE = 6.5
			mainBlock = "3 blocos em ritmo sustentável, com recuperação leve"
			summary = "A meta de prova orienta um estímulo sustentável, sem simular a prova inteira."
		} else if input.ExperienceLevel == "advanced" && input.Cycling.UsesPower && input.Cycling.FTP != nil {
			name = "Sweet spot por potência"
			targetRPE = 7.0
			mainBlock = "3 blocos sustentados guiados pelo FTP informado, com recuperação leve"
			summary = "O medidor de potência e o FTP informados permitem orientar um esforço sustentável."
		} else if input.ExperienceLevel == "advanced" {
			name = "Sweet spot progressivo"
			targetRPE = 7.0
		}
		if input.ExperienceLevel == "advanced" && input.RotationIndex%2 == 1 {
			switch name {
			case "Intervalos controlados":
				name = "Sweet spot progressivo"
				mainBlock = "3 blocos sustentados com recuperação leve"
				summary = "A rotação entre ciclos alterna o estímulo intervalado por um esforço sustentável de qualidade."
				usesControlledIntervals = false
				rotationApplied = true
			case "Subidas controladas", "Sweet spot por potência", "Sweet spot progressivo":
				name = "Tempo controlado"
				targetRPE = 6.0
				mainBlock = "3 blocos sustentados com recuperação leve"
				summary = "A rotação entre ciclos alterna o estímulo específico sem aumentar a carga planejada."
				rotationApplied = true
			}
		}
	}
	if restricted {
		name = "Giro leve protegido"
		targetRPE = 3.5
		mainBlock = "Esforço leve; interromper diante de dor ou desconforto"
		if baseMinutes > 45 {
			baseMinutes = 45
		}
		multiplier *= 0.8
		summary = explanationFor(kind, true)
	}
	duration := int(float64(baseMinutes) * multiplier)
	if duration < 20 {
		duration = 20
	}
	if duration > slot.AvailableMinutes {
		duration = slot.AvailableMinutes
	}

	rules := []string{
		fmt.Sprintf("Agendado em um dia com %d minutos disponíveis.", slot.AvailableMinutes),
		fmt.Sprintf("Carga compatível com experiência %s.", experienceLabel(input.ExperienceLevel)),
		"Progressão de três semanas seguida por uma semana de recuperação.",
	}
	if usesControlledIntervals {
		rules = append(rules, "Intervalos liberados pela avaliação submáxima apta, apenas nas semanas de construção.")
	}
	if rotationApplied {
		rules = append(rules, "Sessão alternada pela rotação explicável entre ciclos.")
	}
	if restricted {
		rules = append(rules, "Intensidade limitada por uma condição de segurança ativa.")
	}
	evidenceKeys := []string{"acsm-1998"}
	if usesControlledIntervals {
		evidenceKeys = append(evidenceKeys, "rosenblat-2020")
	}
	return Workout{
		ScheduledOn:     date.Format("2006-01-02"),
		Name:            name,
		Objective:       objectiveFor(input.PrimaryGoal),
		DurationMinutes: duration,
		TargetRPE:       targetRPE,
		Structure: map[string]any{
			"warmup_minutes":   minInt(10, maxInt(5, duration/6)),
			"main":             mainBlock,
			"cooldown_minutes": minInt(10, maxInt(5, duration/8)),
		},
		Explanation: map[string]any{"summary": summary, "rules": rules, "evidence_keys": evidenceKeys},
		Status:      "planned",
	}
}

func nextMonday(value time.Time) time.Time {
	local := time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
	if local.Weekday() == time.Sunday {
		return local.AddDate(0, 0, 1)
	}
	return local.AddDate(0, 0, -((int(local.Weekday()) + 6) % 7))
}

func weekdayOffset(weekday int) int { return (weekday + 6) % 7 }

func longestSlot(slots []AvailabilitySlot) int {
	longest := 0
	for index := range slots {
		if slots[index].AvailableMinutes > slots[longest].AvailableMinutes {
			longest = index
		}
	}
	return longest
}

func intensitySlot(slots []AvailabilitySlot, longIndex int) int {
	for index := range slots {
		if index != longIndex {
			return index
		}
	}
	return -1
}

func objectiveFor(goal string) string {
	labels := map[string]string{
		"health":            "Construir saúde cardiovascular com consistência",
		"fitness":           "Desenvolver condicionamento geral",
		"endurance":         "Aumentar resistência para pedais mais longos",
		"performance":       "Elevar a capacidade de sustentar esforço",
		"event":             "Criar base específica para o evento-alvo",
		"weight_management": "Apoiar gasto energético com carga sustentável",
	}
	return labels[goal]
}

func explanationFor(kind string, restricted bool) string {
	if restricted {
		return "Sessão deliberadamente leve para respeitar a limitação informada."
	}
	if kind == "long" {
		return "O maior período disponível da semana recebe o estímulo principal de resistência."
	}
	if kind == "quality" {
		return "Uma única sessão de qualidade oferece estímulo sem concentrar carga excessiva."
	}
	return "Sessão de base para acumular consistência com baixo custo de recuperação."
}

func experienceLabel(value string) string {
	return map[string]string{"beginner": "iniciante", "intermediate": "intermediária", "advanced": "avançada"}[value]
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
