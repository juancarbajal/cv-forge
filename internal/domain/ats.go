package domain

import (
	"math"
	"time"
)

type MatchStatus string

const (
	MatchStatusFulfilled MatchStatus = "Cumple"
	MatchStatusPartial   MatchStatus = "Cumple parcialmente"
	MatchStatusMissing   MatchStatus = "No evidenciado"
)

type MatchResult struct {
	ID               string      `json:"id"`
	RequirementID    string      `json:"requisito_id"`
	JobApplicationID string      `json:"job_application_id"`
	RequirementName  string      `json:"requisito"`
	Status           MatchStatus `json:"estado"`
	Confidence       float64     `json:"confianza"` // 0.0 a 1.0
	Evidence         string      `json:"evidencia"`
	Explanation      string      `json:"explicacion"`
}

type ATSScore struct {
	ID                   string          `json:"id"`
	JobApplicationID     string          `json:"job_application_id"`
	HardSkillsScore      float64         `json:"hard_skills_score"`
	ExperienceScore      float64         `json:"experience_score"`
	KeywordsScore        float64         `json:"keywords_score"`
	EducationScore       float64         `json:"education_score"`
	LanguageScore        float64         `json:"language_score"`
	TotalScore           float64         `json:"total_score"`
	RecommendationsJSON  string          `json:"recommendations_json"`
	EvaluatedAt          time.Time       `json:"evaluated_at"`
}

type Recommendations struct {
	Improvements []string `json:"mejoras_recomendadas"`
	Risks        []string `json:"riesgos"`
	Opportunities[]string `json:"oportunidades"`
	Missing      []string `json:"faltantes"`
	Keywords     []string `json:"keywords_ats"`
}

type OptimizationChange struct {
	ID                     string  `json:"id"`
	JobApplicationID       string  `json:"job_application_id"`
	CVVersionGeneratedID   string  `json:"cv_version_generated_id"`
	ChangeSummary          string  `json:"resumen_cambio"`
	Reason                 string  `json:"motivo"`
	RelatedRequirementName string  `json:"requisito_relacionado"`
	ExpectedATSImpact      float64 `json:"impacto_ats_esperado"`
	EvidenceUsed           string  `json:"evidencia_origen"`
}

type OptimizedCVResult struct {
	ContentMarkdown string               `json:"content_markdown"`
	Changes         []OptimizationChange `json:"changes"`
	TargetLanguage  Language             `json:"target_language"`
}

// Pesos fijos ATS según RF-017
const (
	WeightHardSkills float64 = 0.35
	WeightExperience float64 = 0.25
	WeightKeywords   float64 = 0.20
	WeightEducation  float64 = 0.10
	WeightLanguage   float64 = 0.10
)

// CalculateATSScore calcula el puntaje ATS de forma explicable y determinista
func CalculateATSScore(jobAppID string, reqs []Requirement, matches []MatchResult) ATSScore {
	reqMap := make(map[string]Requirement)
	for _, r := range reqs {
		reqMap[r.ID] = r
		if r.Name != "" {
			reqMap[r.Name] = r
		}
	}

	categoryPoints := make(map[RequirementCategory]float64)
	categoryMaxPoints := make(map[RequirementCategory]float64)

	// Inicializar categorías base
	cats := []RequirementCategory{CategoryHardSkill, CategoryExperience, CategoryKeywords, CategoryEducation, CategoryLanguage}
	for _, c := range cats {
		categoryPoints[c] = 0
		categoryMaxPoints[c] = 0
	}

	for _, m := range matches {
		req, exists := reqMap[m.RequirementID]
		if !exists {
			req, exists = reqMap[m.RequirementName]
		}
		if !exists {
			continue
		}

		// Ponderación por prioridad
		priorityWeight := 1.0
		switch req.Priority {
		case PriorityMandatory:
			priorityWeight = 2.0
		case PriorityDesirable:
			priorityWeight = 1.0
		case PriorityOptional:
			priorityWeight = 0.5
		}

		categoryMaxPoints[req.Category] += priorityWeight

		switch m.Status {
		case MatchStatusFulfilled:
			categoryPoints[req.Category] += priorityWeight
		case MatchStatusPartial:
			conf := m.Confidence
			if conf <= 0 {
				conf = 0.5
			} else if conf > 1.0 {
				conf = 1.0
			}
			categoryPoints[req.Category] += priorityWeight * conf
		case MatchStatusMissing:
			// 0 puntos
		}
	}

	calcCatScore := func(cat RequirementCategory) float64 {
		max := categoryMaxPoints[cat]
		if max <= 0 {
			// Si la oferta no especifica requisitos de esta categoría, asigna 100 por defecto
			return 100.0
		}
		pts := categoryPoints[cat]
		score := (pts / max) * 100.0
		if score > 100.0 {
			score = 100.0
		}
		return math.Round(score*100) / 100
	}

	hardScore := calcCatScore(CategoryHardSkill)
	expScore := calcCatScore(CategoryExperience)
	kwScore := calcCatScore(CategoryKeywords)
	eduScore := calcCatScore(CategoryEducation)
	langScore := calcCatScore(CategoryLanguage)

	total := (hardScore * WeightHardSkills) +
		(expScore * WeightExperience) +
		(kwScore * WeightKeywords) +
		(eduScore * WeightEducation) +
		(langScore * WeightLanguage)

	total = math.Round(total*100) / 100

	return ATSScore{
		JobApplicationID: jobAppID,
		HardSkillsScore:  hardScore,
		ExperienceScore:  expScore,
		KeywordsScore:    kwScore,
		EducationScore:   eduScore,
		LanguageScore:    langScore,
		TotalScore:       total,
		EvaluatedAt:      time.Now().UTC(),
	}
}
