package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jcarbajal/cv-checker/internal/domain"
)

// Prompts por defecto que se siembran en la base de datos (SeedDefaults)
var DefaultPrompts = []domain.Prompt{
	{
		Type:    domain.PromptTypeJobAnalysis,
		Name:    "Extracción y Clasificación de Requisitos",
		Version: 1,
		Content: `Eres un asistente experto en selección técnica y sistemas ATS.
Analiza la siguiente oferta laboral y extrae los requisitos explícitos e implícitos.
Clasifica cada requisito en una categoría: 'hard_skill', 'experience', 'keywords', 'education', 'language'.
Asigna una prioridad a cada uno: 'obligatorio', 'deseable', 'opcional'.

Responde ÚNICAMENTE un objeto JSON válido con la siguiente estructura:
{
  "detected_language": "es|en|pt",
  "requirements": [
    {
      "name": "Nombre conciso de la habilidad o requisito",
      "categoria": "hard_skill|experience|keywords|education|language",
      "prioridad": "obligatorio|deseable|opcional"
    }
  ]
}`,
	},
	{
		Type:    domain.PromptTypeCVMatching,
		Name:    "Comparación CV vs Requisitos con Evidencia",
		Version: 1,
		Content: `Eres un auditor ATS estricto y objetivo.
Compara el CV proporcionado contra la lista de requisitos de la oferta laboral.
Para cada requisito, determina si el CV:
- 'Cumple': hay evidencia directa o sinónimos claros (ej: Go = Golang, Postgres = PostgreSQL, K8s = Kubernetes).
- 'Cumple parcialmente': hay experiencia cercana o tecnologías relacionadas sin mención explícita.
- 'No evidenciado': no hay ninguna mención ni indicio en el CV.

Proporciona el nivel de confianza (entre 0.0 y 1.0), cita la evidencia textual encontrada y una breve explicación.

Responde ÚNICAMENTE un objeto JSON válido con la siguiente estructura:
{
  "matches": [
    {
      "requisito": "Nombre del requisito evaluado",
      "estado": "Cumple|Cumple parcialmente|No evidenciado",
      "confianza": 0.95,
      "evidencia": "Cita o referencia encontrada en el CV",
      "explicacion": "Razón de la clasificación"
    }
  ]
}`,
	},
	{
		Type:    domain.PromptTypeATSScoring,
		Name:    "Recomendaciones y Keywords ATS",
		Version: 1,
		Content: `Eres un consultor de carrera y optimización para filtros ATS.
Con base en los requisitos y el matching del CV, genera recomendaciones de mejora:
- Mejoras recomendadas (bullet points cuantificables).
- Riesgos (aspectos débiles frente a la vacante).
- Oportunidades.
- Faltantes.
- Keywords ATS más relevantes.

Responde ÚNICAMENTE un objeto JSON válido:
{
  "mejoras_recomendadas": ["..."],
  "riesgos": ["..."],
  "oportunidades": ["..."],
  "faltantes": ["..."],
  "keywords_ats": ["..."]
}`,
	},
	{
		Type:    domain.PromptTypeCVOptimization,
		Name:    "Optimización de CV y Trazabilidad Anti-Invención",
		Version: 1,
		Content: `Eres un redactor profesional de currículums técnicos y experto en optimización ATS.
Tu objetivo es reescribir y optimizar el CV Markdown adaptándolo a la oferta laboral en el idioma objetivo indicado.

REGLAS ESTRICTAS ANTI-INVENCIÓN:
- NO inventes experiencia laboral, empresas, cargos o fechas.
- NO inventes tecnologías que el candidato no conozca ni mencione.
- Toda mejora debe derivarse de la evidencia preexistente en el CV original (enfatizar logros, reorganizar habilidades, usar sinónimos estándar de la industria).
- Para cada cambio relevante realizado, debes reportar el motivo, el requisito relacionado y la evidencia original que lo sustenta.

Responde ÚNICAMENTE un objeto JSON válido con la siguiente estructura:
{
  "content_markdown": "# Contenido completo del CV optimizado en Markdown...",
  "changes": [
    {
      "resumen_cambio": "Descripción breve del cambio",
      "motivo": "Motivación del cambio en relación al anuncio",
      "requisito_relacionado": "Nombre del requisito impactado",
      "impacto_ats_esperado": 3.5,
      "evidencia_origen": "Evidencia existente en el CV original que sustenta este cambio"
    }
  ]
}`,
	},
	{
		Type:    domain.PromptTypeTranslation,
		Name:    "Traducción de CV Preservando Contexto",
		Version: 1,
		Content: `Eres un traductor profesional de perfiles técnicos de ingeniería.
Traduce el siguiente CV al idioma objetivo indicado (Español, Inglés o Portugués), preservando la terminología técnica estándar de la industria, las viñetas y el formato Markdown.`,
	},
	{
		Type:    domain.PromptTypeCriticAudit,
		Name:    "Auditor Semántico Anti-Invención (Critic Agent)",
		Version: 1,
		Content: `Eres un auditor estricto de veracidad curricular (Critic Agent).
Compara el CV ORIGINAL (y las respuestas verificadas del candidato, si existen) contra el CV OPTIMIZADO y su lista de cambios propuestos.
Detecta si se ha exagerado el seniority, inventado métricas cuantitativas sin respaldo, añadido tecnologías no presentes en la evidencia o distorsionado cargos/fechas.

Responde ÚNICAMENTE un objeto JSON válido con la siguiente estructura:
{
  "approved": true,
  "issues": [
    {
      "severity": "WARNING|BLOCKING",
      "claim": "Afirmación cuestionada en el CV optimizado",
      "reason": "Por qué no está suficientemente respaldada por el CV original",
      "suggestion": "Cómo reformularla o eliminarla"
    }
  ],
  "summary": "Resumen breve del veredicto de auditoría"
}`,
	},
	{
		Type:    domain.PromptTypeGapInterview,
		Name:    "Entrevistador de Gaps Curriculares (Gap Interviewer Agent)",
		Version: 1,
		Content: `Eres un entrevistador técnico experto que ayuda a candidatos a descubrir experiencia real no documentada en su CV.
Analiza los requisitos con estado 'No evidenciado' o 'Cumple parcialmente' y genera hasta 5 preguntas concretas y directas para confirmar si el candidato posee esa experiencia en proyectos reales.

Responde ÚNICAMENTE un objeto JSON válido con la siguiente estructura:
{
  "questions": [
    {
      "requirement_name": "Nombre del requisito",
      "question": "Pregunta específica para el candidato",
      "why_asked": "Breve explicación de por qué esta información ayudaría al matching ATS"
    }
  ]
}`,
	},
}

// DTOs auxiliares para deserialización de respuestas JSON de LLMs
type jobAnalysisResponse struct {
	DetectedLanguage string `json:"detected_language"`
	Requirements     []struct {
		Name     string `json:"name"`
		Category string `json:"categoria"`
		Priority string `json:"prioridad"`
	} `json:"requirements"`
}

type cvMatchingResponse struct {
	Matches []struct {
		Requirement string  `json:"requisito"`
		Status      string  `json:"estado"`
		Confidence  float64 `json:"confianza"`
		Evidence    string  `json:"evidencia"`
		Explanation string  `json:"explicacion"`
	} `json:"matches"`
}

type recommendationsResponse struct {
	Improvements  []string `json:"mejoras_recomendadas"`
	Risks         []string `json:"riesgos"`
	Opportunities []string `json:"oportunidades"`
	Missing       []string `json:"faltantes"`
	Keywords      []string `json:"keywords_ats"`
}

type cvOptimizationResponse struct {
	ContentMarkdown string `json:"content_markdown"`
	Changes         []struct {
		Summary           string  `json:"resumen_cambio"`
		Reason            string  `json:"motivo"`
		RelatedReq        string  `json:"requisito_relacionado"`
		ExpectedATSImpact float64 `json:"impacto_ats_esperado"`
		EvidenceUsed      string  `json:"evidencia_origen"`
	} `json:"changes"`
}

// CleanJSONResponse limpia delimitadores de bloques de código markdown como ```json ... ```
func CleanJSONResponse(raw string) string {
	cleaned := strings.TrimSpace(raw)
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimSuffix(cleaned, "```")
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
	}
	return strings.TrimSpace(cleaned)
}

func parseRequirements(jsonStr string) (domain.Language, []domain.Requirement, error) {
	cleaned := CleanJSONResponse(jsonStr)
	var resp jobAnalysisResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return domain.LangSpanish, nil, fmt.Errorf("error parseando requisitos: %w (raw: %s)", err, jsonStr)
	}

	lang := domain.Language(resp.DetectedLanguage)
	if lang == "" {
		lang = domain.LangSpanish
	}

	var reqs []domain.Requirement
	for _, r := range resp.Requirements {
		reqs = append(reqs, domain.Requirement{
			Name:     r.Name,
			Category: domain.RequirementCategory(r.Category),
			Priority: domain.RequirementPriority(r.Priority),
		})
	}
	return lang, reqs, nil
}

func parseMatches(jsonStr string) ([]domain.MatchResult, error) {
	cleaned := CleanJSONResponse(jsonStr)
	var resp cvMatchingResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return nil, fmt.Errorf("error parseando matching: %w", err)
	}

	var matches []domain.MatchResult
	for _, m := range resp.Matches {
		matches = append(matches, domain.MatchResult{
			RequirementName: m.Requirement,
			Status:          domain.MatchStatus(m.Status),
			Confidence:      m.Confidence,
			Evidence:        m.Evidence,
			Explanation:     m.Explanation,
		})
	}
	return matches, nil
}

func parseRecommendations(jsonStr string) (domain.Recommendations, error) {
	cleaned := CleanJSONResponse(jsonStr)
	var resp recommendationsResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return domain.Recommendations{}, fmt.Errorf("error parseando recomendaciones: %w", err)
	}
	return domain.Recommendations{
		Improvements:  resp.Improvements,
		Risks:         resp.Risks,
		Opportunities: resp.Opportunities,
		Missing:       resp.Missing,
		Keywords:      resp.Keywords,
	}, nil
}

func parseOptimization(jsonStr string, targetLang domain.Language) (domain.OptimizedCVResult, error) {
	cleaned := CleanJSONResponse(jsonStr)
	var resp cvOptimizationResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return domain.OptimizedCVResult{}, fmt.Errorf("error parseando optimización de CV: %w", err)
	}

	var changes []domain.OptimizationChange
	for _, c := range resp.Changes {
		changes = append(changes, domain.OptimizationChange{
			ChangeSummary:          c.Summary,
			Reason:                 c.Reason,
			RelatedRequirementName: c.RelatedReq,
			ExpectedATSImpact:      c.ExpectedATSImpact,
			EvidenceUsed:           c.EvidenceUsed,
		})
	}

	return domain.OptimizedCVResult{
		ContentMarkdown: resp.ContentMarkdown,
		Changes:         changes,
		TargetLanguage:  targetLang,
	}, nil
}
