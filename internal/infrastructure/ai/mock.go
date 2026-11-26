package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/jcarbajal/cv-checker/internal/domain"
)

type MockAIProvider struct {
	Name string
}

func NewMockAIProvider() *MockAIProvider {
	return &MockAIProvider{Name: "mock"}
}

func (m *MockAIProvider) ProviderName() string {
	return m.Name
}

func (m *MockAIProvider) DetectLanguage(ctx context.Context, text string) (domain.Language, error) {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "experiência") || strings.Contains(lower, "vaga") {
		return domain.LangPortuguese, nil
	}
	if strings.Contains(lower, "experience") || strings.Contains(lower, "looking for") || strings.Contains(lower, "requirements") {
		return domain.LangEnglish, nil
	}
	return domain.LangSpanish, nil
}

func (m *MockAIProvider) ExtractAndClassifyRequirements(ctx context.Context, jobPosting string, prompt string) ([]domain.Requirement, error) {
	// Requisitos simulados basados en contenido de la oferta
	reqs := []domain.Requirement{
		{
			Name:     "Go / Golang",
			Category: domain.CategoryHardSkill,
			Priority: domain.PriorityMandatory,
		},
		{
			Name:     "PostgreSQL",
			Category: domain.CategoryHardSkill,
			Priority: domain.PriorityDesirable,
		},
		{
			Name:     "Docker & Contenedores",
			Category: domain.CategoryHardSkill,
			Priority: domain.PriorityMandatory,
		},
		{
			Name:     "Experiencia Backend 3+ años",
			Category: domain.CategoryExperience,
			Priority: domain.PriorityMandatory,
		},
		{
			Name:     "Inglés Técnico",
			Category: domain.CategoryLanguage,
			Priority: domain.PriorityDesirable,
		},
	}
	return reqs, nil
}

func (m *MockAIProvider) MatchCV(ctx context.Context, cvContent string, requirements []domain.Requirement, prompt string) ([]domain.MatchResult, error) {
	cvLower := strings.ToLower(cvContent)
	var matches []domain.MatchResult

	for _, r := range requirements {
		reqName := r.Name
		reqLower := strings.ToLower(reqName)

		status := domain.MatchStatusMissing
		confidence := 0.0
		evidence := ""
		explanation := "No se detectó mención explícita ni relacionada en el CV."

		if strings.Contains(cvLower, "go") || strings.Contains(cvLower, "golang") {
			if strings.Contains(reqLower, "go") {
				status = domain.MatchStatusFulfilled
				confidence = 0.98
				evidence = "Mención explícita de proyectos y desarrollo en Go."
				explanation = "Cumple plenamente con el requisito técnico."
			}
		}
		if strings.Contains(cvLower, "postgres") || strings.Contains(cvLower, "sql") {
			if strings.Contains(reqLower, "postgre") {
				status = domain.MatchStatusFulfilled
				confidence = 0.95
				evidence = "Experiencia con bases de datos SQL y PostgreSQL."
				explanation = "Evidencia clara de uso en producción."
			}
		}
		if strings.Contains(cvLower, "docker") || strings.Contains(cvLower, "contenedores") || strings.Contains(cvLower, "linux") {
			if strings.Contains(reqLower, "docker") {
				status = domain.MatchStatusPartial
				confidence = 0.84
				evidence = "Experiencia con contenedores Linux."
				explanation = "Conoce tecnologías de contenedorización afines."
			}
		}
		if strings.Contains(cvLower, "backend") {
			if strings.Contains(reqLower, "backend") {
				status = domain.MatchStatusFulfilled
				confidence = 1.0
				evidence = "Historial profesional en desarrollo backend."
				explanation = "Trayectoria comprobable."
			}
		}

		matches = append(matches, domain.MatchResult{
			RequirementName: reqName,
			Status:          status,
			Confidence:      confidence,
			Evidence:        evidence,
			Explanation:     explanation,
		})
	}
	return matches, nil
}

func (m *MockAIProvider) GenerateRecommendations(ctx context.Context, cvContent string, matches []domain.MatchResult, prompt string) (domain.Recommendations, error) {
	return domain.Recommendations{
		Improvements: []string{
			"Destacar métricas cuantitativas de impacto en los proyectos de backend.",
			"Especificar el uso de Docker y orquestación para aumentar el score de DevOps.",
		},
		Risks: []string{
			"Falta mención formal de certificación de inglés si la vacante es internacional.",
		},
		Opportunities: []string{
			"Alinear la sección de habilidades técnicas al formato estándar ATS.",
		},
		Missing: []string{
			"Mención de despliegues en Kubernetes o nubes públicas.",
		},
		Keywords: []string{
			"Golang", "Microservicios", "PostgreSQL", "Docker", "REST API",
		},
	}, nil
}

func (m *MockAIProvider) GenerateOptimizedCV(ctx context.Context, cvContent string, jobPosting string, matches []domain.MatchResult, targetLang domain.Language, prompt string) (domain.OptimizedCVResult, error) {
	optimizedContent := fmt.Sprintf("%s\n\n## Habilidades Técnicas Optimizadas\n- **Lenguajes:** Go (Golang), SQL\n- **Bases de Datos:** PostgreSQL\n- **Infraestructura:** Docker y Contenedores Linux\n", cvContent)

	changes := []domain.OptimizationChange{
		{
			ChangeSummary:          "Se reorganizó la sección técnica incorporando PostgreSQL y Go como keywords principales",
			Reason:                 "Requisitos obligatorios y deseables detectados en la oferta laboral",
			RelatedRequirementName: "Go / Golang",
			ExpectedATSImpact:      4.5,
			EvidenceUsed:           "El CV original mencionaba desarrollo backend y bases de datos relacionales",
		},
		{
			ChangeSummary:          "Se enfatizó la experiencia en contenedores y despliegues",
			Reason:                 "Requisito obligatorio de Docker en la vacante",
			RelatedRequirementName: "Docker & Contenedores",
			ExpectedATSImpact:      3.0,
			EvidenceUsed:           "El CV citaba contenedores Linux en proyectos anteriores",
		},
	}

	return domain.OptimizedCVResult{
		ContentMarkdown: optimizedContent,
		Changes:         changes,
		TargetLanguage:  targetLang,
	}, nil
}

func (m *MockAIProvider) CompleteJSON(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	lowerSys := strings.ToLower(systemPrompt)
	if strings.Contains(lowerSys, "auditor") || strings.Contains(lowerSys, "approved") {
		return `{"approved": true, "issues": [], "summary": "Todos los cambios están respaldados por evidencia del CV original."}`, nil
	}
	if strings.Contains(lowerSys, "entrevistador") || strings.Contains(lowerSys, "questions") {
		return `{"questions": [{"requirement_name": "PostgreSQL", "question": "¿Has administrado o diseñado esquemas en PostgreSQL en tus proyectos backend?", "why_asked": "El CV menciona bases de datos relacionales sin detallar PostgreSQL."}]}`, nil
	}
	return `{}`, nil
}
