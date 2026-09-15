package athlete

import (
	"fmt"
	"strings"
)

const QuestionnaireVersion = "cycling-onboarding-v2"

type QuestionCondition struct {
	QuestionID string `json:"question_id"`
	Equals     any    `json:"equals"`
}

type QuestionDefinition struct {
	ID        string             `json:"id"`
	Step      int                `json:"step"`
	Prompt    string             `json:"prompt"`
	Kind      string             `json:"kind"`
	Required  bool               `json:"required"`
	Condition *QuestionCondition `json:"condition,omitempty"`
}

type QuestionnaireStep struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	QuestionIDs []string `json:"question_ids"`
}

type QuestionnaireDefinition struct {
	Version   string               `json:"version"`
	Scope     string               `json:"scope"`
	Steps     []QuestionnaireStep  `json:"steps"`
	Questions []QuestionDefinition `json:"questions"`
}

// CyclingQuestionnaire is the source of truth for order and conditionality.
// The UI can keep specialized controls while other clients consume the same
// question -> answer -> rule -> next question contract.
func CyclingQuestionnaire() QuestionnaireDefinition {
	return QuestionnaireDefinition{
		Version: QuestionnaireVersion,
		Scope:   "cycling_profile_and_plan_generation",
		Steps: []QuestionnaireStep{
			{ID: "profile", Title: "Conte-nos onde você está agora.", Description: "Esses dados definem os limites iniciais.", QuestionIDs: []string{"birth_date", "sex", "height_cm", "weight_kg", "experience_level", "activity_level", "waist_cm", "body_fat_percent", "weight_trend"}},
			{ID: "safety", Title: "Existe algo que o treino deve respeitar?", Description: "Dor, lesões e limitações têm prioridade sobre desempenho.", QuestionIDs: []string{"has_limitation", "limitation_kind", "limitation_description", "recent_surgery", "exercise_prohibited", "condition_affecting_exercise"}},
			{ID: "objective", Title: "Onde você quer chegar?", Description: "Defina uma direção principal e, se desejar, uma prioridade secundária.", QuestionIDs: []string{"primary_goal", "secondary_goal", "target_date", "goal_details"}},
			{ID: "routine", Title: "Quanto tempo cabe na sua semana?", Description: "A rotina informada orienta a distribuição dos treinos.", QuestionIDs: []string{"discipline", "practice_duration_months", "weekly_hours", "weekly_rides", "recent_weekly_distance_km", "average_ride_minutes", "longest_ride_minutes", "training_status", "recent_training_weeks", "recent_best_distance_km", "bike_type", "terrain", "uses_gps", "uses_sports_watch", "uses_smart_trainer", "uses_heart_rate", "uses_power", "ftp", "ftp_test_date", "ftp_protocol", "average_power_watts", "event_goal", "event_distance_km", "event_date", "available_days", "preferred_time", "location"}},
		},
		Questions: []QuestionDefinition{
			{ID: "birth_date", Step: 1, Prompt: "Data de nascimento", Kind: "date"},
			{ID: "sex", Step: 1, Prompt: "Sexo", Kind: "choice"},
			{ID: "height_cm", Step: 1, Prompt: "Altura", Kind: "number"},
			{ID: "weight_kg", Step: 1, Prompt: "Peso atual", Kind: "number"},
			{ID: "experience_level", Step: 1, Prompt: "Experiência no ciclismo", Kind: "choice", Required: true},
			{ID: "activity_level", Step: 1, Prompt: "Rotina de atividade atual", Kind: "choice"},
			{ID: "waist_cm", Step: 1, Prompt: "Circunferência da cintura", Kind: "number"},
			{ID: "body_fat_percent", Step: 1, Prompt: "Composição corporal", Kind: "number"},
			{ID: "weight_trend", Step: 1, Prompt: "Histórico recente do peso", Kind: "choice"},
			{ID: "has_limitation", Step: 2, Prompt: "Existe uma limitação atual?", Kind: "boolean", Required: true},
			{ID: "limitation_kind", Step: 2, Prompt: "Tipo de limitação", Kind: "choice", Condition: &QuestionCondition{QuestionID: "has_limitation", Equals: true}},
			{ID: "limitation_description", Step: 2, Prompt: "Descrição da limitação", Kind: "text", Condition: &QuestionCondition{QuestionID: "has_limitation", Equals: true}},
			{ID: "recent_surgery", Step: 2, Prompt: "Houve cirurgia recente?", Kind: "boolean", Condition: &QuestionCondition{QuestionID: "has_limitation", Equals: true}},
			{ID: "exercise_prohibited", Step: 2, Prompt: "Existe proibição de exercício?", Kind: "boolean", Condition: &QuestionCondition{QuestionID: "has_limitation", Equals: true}},
			{ID: "condition_affecting_exercise", Step: 2, Prompt: "Existe condição que afeta o exercício?", Kind: "boolean", Condition: &QuestionCondition{QuestionID: "has_limitation", Equals: true}},
			{ID: "primary_goal", Step: 3, Prompt: "Objetivo principal", Kind: "choice", Required: true},
			{ID: "secondary_goal", Step: 3, Prompt: "Objetivo secundário", Kind: "choice"},
			{ID: "target_date", Step: 3, Prompt: "Data-alvo", Kind: "date"},
			{ID: "goal_details", Step: 3, Prompt: "Detalhes do objetivo", Kind: "text"},
			{ID: "discipline", Step: 4, Prompt: "Modalidade principal", Kind: "choice"},
			{ID: "practice_duration_months", Step: 4, Prompt: "Tempo praticando ciclismo", Kind: "number"},
			{ID: "weekly_hours", Step: 4, Prompt: "Horas por semana", Kind: "number"},
			{ID: "weekly_rides", Step: 4, Prompt: "Pedais por semana", Kind: "number"},
			{ID: "recent_weekly_distance_km", Step: 4, Prompt: "Distância semanal recente", Kind: "number"},
			{ID: "average_ride_minutes", Step: 4, Prompt: "Duração média do pedal", Kind: "number"},
			{ID: "longest_ride_minutes", Step: 4, Prompt: "Maior pedal recente", Kind: "number"},
			{ID: "training_status", Step: 4, Prompt: "Situação atual do treino", Kind: "choice"},
			{ID: "recent_training_weeks", Step: 4, Prompt: "Semanas treinando com regularidade", Kind: "number"},
			{ID: "recent_best_distance_km", Step: 4, Prompt: "Maior distância recente", Kind: "number"},
			{ID: "bike_type", Step: 4, Prompt: "Tipo de bicicleta", Kind: "choice"},
			{ID: "terrain", Step: 4, Prompt: "Terreno predominante", Kind: "choice"},
			{ID: "uses_gps", Step: 4, Prompt: "Uso GPS", Kind: "boolean"},
			{ID: "uses_sports_watch", Step: 4, Prompt: "Uso relógio esportivo", Kind: "boolean"},
			{ID: "uses_smart_trainer", Step: 4, Prompt: "Uso rolo inteligente", Kind: "boolean"},
			{ID: "uses_heart_rate", Step: 4, Prompt: "Uso frequência cardíaca", Kind: "boolean"},
			{ID: "uses_power", Step: 4, Prompt: "Uso medidor de potência", Kind: "boolean"},
			{ID: "ftp", Step: 4, Prompt: "FTP", Kind: "number", Condition: &QuestionCondition{QuestionID: "uses_power", Equals: true}},
			{ID: "ftp_test_date", Step: 4, Prompt: "Data do teste de FTP", Kind: "date", Condition: &QuestionCondition{QuestionID: "uses_power", Equals: true}},
			{ID: "ftp_protocol", Step: 4, Prompt: "Protocolo do teste de FTP", Kind: "choice", Condition: &QuestionCondition{QuestionID: "uses_power", Equals: true}},
			{ID: "average_power_watts", Step: 4, Prompt: "Potência média", Kind: "number", Condition: &QuestionCondition{QuestionID: "uses_power", Equals: true}},
			{ID: "event_goal", Step: 4, Prompt: "Preparação para um evento", Kind: "boolean"},
			{ID: "event_distance_km", Step: 4, Prompt: "Distância do evento", Kind: "number", Condition: &QuestionCondition{QuestionID: "event_goal", Equals: true}},
			{ID: "event_date", Step: 4, Prompt: "Data do evento", Kind: "date", Condition: &QuestionCondition{QuestionID: "event_goal", Equals: true}},
			{ID: "available_days", Step: 4, Prompt: "Dias disponíveis", Kind: "availability", Required: true},
			{ID: "preferred_time", Step: 4, Prompt: "Horário preferido", Kind: "time"},
			{ID: "location", Step: 4, Prompt: "Local disponível", Kind: "choice"},
		},
	}
}

// VisibleQuestions returns the questions whose conditional gates are satisfied.
// Clients can use it to render an adaptive form without duplicating gates.
func VisibleQuestions(definition QuestionnaireDefinition, answers map[string]any) []QuestionDefinition {
	visible := make([]QuestionDefinition, 0, len(definition.Questions))
	for _, question := range definition.Questions {
		if question.Condition != nil && !conditionMatches(question.Condition, answers) {
			continue
		}
		visible = append(visible, question)
	}
	return visible
}

// NextQuestion returns the first unanswered required question whose condition
// is true. Optional questions remain visible, but never block completion.
func NextQuestion(definition QuestionnaireDefinition, answers map[string]any) string {
	for _, question := range VisibleQuestions(definition, answers) {
		if !question.Required {
			continue
		}
		if value, ok := answers[question.ID]; !ok || strings.TrimSpace(toAnswerString(value)) == "" {
			return question.ID
		}
	}
	return ""
}

func conditionMatches(condition *QuestionCondition, answers map[string]any) bool {
	value, ok := answers[condition.QuestionID]
	if !ok {
		return false
	}
	return toAnswerString(value) == toAnswerString(condition.Equals)
}

func toAnswerString(value any) string {
	switch typed := value.(type) {
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case string:
		return strings.TrimSpace(typed)
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}
