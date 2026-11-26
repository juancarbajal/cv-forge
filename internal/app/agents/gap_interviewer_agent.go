package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jcarbajal/cv-checker/internal/domain"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/ai"
)

// GapInterviewerAgent genera preguntas puntuales para rescatar experiencia real no documentada en el CV.
type GapInterviewerAgent struct {
	provider     domain.AIProvider
	systemPrompt string
}

// NewGapInterviewerAgent crea un nuevo agente entrevistador de gaps.
func NewGapInterviewerAgent(provider domain.AIProvider, systemPrompt string) *GapInterviewerAgent {
	if strings.TrimSpace(systemPrompt) == "" {
		for _, p := range ai.DefaultPrompts {
			if p.Type == domain.PromptTypeGapInterview {
				systemPrompt = p.Content
				break
			}
		}
	}
	return &GapInterviewerAgent{
		provider:     provider,
		systemPrompt: systemPrompt,
	}
}

type gapQuestionsResponse struct {
	Questions []domain.GapQuestion `json:"questions"`
}

// GenerateQuestions analiza los requisitos faltantes o parciales y devuelve hasta 5 preguntas concretas.
func (g *GapInterviewerAgent) GenerateQuestions(ctx context.Context, cvContent string, matches []domain.MatchResult) ([]domain.GapQuestion, error) {
	var gaps []domain.MatchResult
	for _, m := range matches {
		if m.Status == domain.MatchStatusMissing || m.Status == domain.MatchStatusPartial {
			gaps = append(gaps, m)
		}
	}
	if len(gaps) == 0 {
		return nil, nil
	}

	gapsJSON, _ := json.Marshal(gaps)
	userMsg := fmt.Sprintf("REQUISITOS CON GAPS O PARCIALES:\n%s\n\nCV DEL CANDIDATO:\n%s", string(gapsJSON), cvContent)

	raw, err := g.provider.CompleteJSON(ctx, g.systemPrompt, userMsg)
	if err != nil {
		return nil, fmt.Errorf("gap interviewer error: %w", err)
	}

	cleaned := ai.CleanJSONResponse(raw)
	var resp gapQuestionsResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return nil, fmt.Errorf("error parseando preguntas de gap interviewer: %w", err)
	}

	if len(resp.Questions) > 5 {
		resp.Questions = resp.Questions[:5]
	}
	return resp.Questions, nil
}

// FormatVerifiedContext convierte las respuestas del candidato en un bloque de evidencia verificada.
func FormatVerifiedContext(answers []domain.GapAnswer) string {
	var b strings.Builder
	for _, ans := range answers {
		trimmed := strings.TrimSpace(ans.Answer)
		if trimmed == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("- Requisito [%s]: %s\n  Respuesta verificada: %s\n", ans.RequirementName, ans.Question, trimmed))
	}
	if b.Len() == 0 {
		return "RESPUESTAS ADICIONALES VERIFICADAS: Ninguna."
	}
	return "RESPUESTAS ADICIONALES VERIFICADAS POR EL CANDIDATO:\n" + b.String()
}
