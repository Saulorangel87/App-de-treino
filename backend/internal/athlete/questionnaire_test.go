package athlete

import "testing"

func TestCyclingQuestionnaireHasOnlyCyclingScope(t *testing.T) {
	definition := CyclingQuestionnaire()
	if definition.Version != QuestionnaireVersion || definition.Scope != "cycling_profile_and_plan_generation" {
		t.Fatalf("unexpected questionnaire contract: %#v", definition)
	}
	if len(definition.Steps) != 4 {
		t.Fatalf("expected four onboarding steps, got %d", len(definition.Steps))
	}
}

func TestVisibleQuestionsHonorsConditionalGates(t *testing.T) {
	definition := CyclingQuestionnaire()
	withoutPower := VisibleQuestions(definition, map[string]any{"uses_power": false})
	withPower := VisibleQuestions(definition, map[string]any{"uses_power": true})
	contains := func(questions []QuestionDefinition, id string) bool {
		for _, question := range questions {
			if question.ID == id {
				return true
			}
		}
		return false
	}
	if contains(withoutPower, "ftp") || !contains(withPower, "ftp") {
		t.Fatalf("power gate was not applied: without=%#v with=%#v", withoutPower, withPower)
	}
}

func TestNextQuestionSkipsOptionalFields(t *testing.T) {
	definition := CyclingQuestionnaire()
	answers := map[string]any{
		"experience_level": "beginner",
		"has_limitation":   false,
		"primary_goal":     "health",
	}
	if next := NextQuestion(definition, answers); next != "available_days" {
		t.Fatalf("expected required availability question, got %q", next)
	}
}
