package domain

import (
	"testing"
)

func TestCalculateATSScore_WeightsAndCategories(t *testing.T) {
	reqs := []Requirement{
		{ID: "req-1", Name: "Go", Category: CategoryHardSkill, Priority: PriorityMandatory},
		{ID: "req-2", Name: "PostgreSQL", Category: CategoryHardSkill, Priority: PriorityDesirable},
		{ID: "req-3", Name: "5 años experiencia backend", Category: CategoryExperience, Priority: PriorityMandatory},
		{ID: "req-4", Name: "REST APIs", Category: CategoryKeywords, Priority: PriorityMandatory},
		{ID: "req-5", Name: "Grado en Informática", Category: CategoryEducation, Priority: PriorityMandatory},
		{ID: "req-6", Name: "Inglés Avanzado", Category: CategoryLanguage, Priority: PriorityMandatory},
	}

	matches := []MatchResult{
		{RequirementID: "req-1", Status: MatchStatusFulfilled},
		{RequirementID: "req-2", Status: MatchStatusPartial, Confidence: 0.5},
		{RequirementID: "req-3", Status: MatchStatusFulfilled},
		{RequirementID: "req-4", Status: MatchStatusFulfilled},
		{RequirementID: "req-5", Status: MatchStatusFulfilled},
		{RequirementID: "req-6", Status: MatchStatusMissing},
	}

	score := CalculateATSScore("job-123", reqs, matches)

	if score.JobApplicationID != "job-123" {
		t.Errorf("expected jobAppID 'job-123', got '%s'", score.JobApplicationID)
	}

	// Hard skills:
	// req-1 (mandatory, weight 2.0): Fulfilled -> 2.0 pts
	// req-2 (desirable, weight 1.0): Partial (0.5) -> 0.5 pts
	// total = 2.5 / 3.0 = 83.33%
	if score.HardSkillsScore < 83.0 || score.HardSkillsScore > 84.0 {
		t.Errorf("unexpected HardSkillsScore: %f", score.HardSkillsScore)
	}

	// Experience:
	// req-3 (mandatory, weight 2.0): Fulfilled -> 2.0/2.0 = 100%
	if score.ExperienceScore != 100.0 {
		t.Errorf("expected ExperienceScore 100, got %f", score.ExperienceScore)
	}

	// Language:
	// req-6 (mandatory, weight 2.0): Missing -> 0%
	if score.LanguageScore != 0.0 {
		t.Errorf("expected LanguageScore 0, got %f", score.LanguageScore)
	}

	// TotalScore must be within 0 and 100
	if score.TotalScore <= 0 || score.TotalScore > 100 {
		t.Errorf("TotalScore out of bounds: %f", score.TotalScore)
	}
}
