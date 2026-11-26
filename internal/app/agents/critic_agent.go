package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jcarbajal/cv-checker/internal/domain"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/ai"
)

// CriticAgent audita semánticamente un CV optimizado frente al CV original y respuestas verificadas.
type CriticAgent struct {
	provider     domain.AIProvider
	systemPrompt string
}

// NewCriticAgent crea un nuevo agente auditor de veracidad curricular.
func NewCriticAgent(provider domain.AIProvider, systemPrompt string) *CriticAgent {
	if strings.TrimSpace(systemPrompt) == "" {
		for _, p := range ai.DefaultPrompts {
			if p.Type == domain.PromptTypeCriticAudit {
				systemPrompt = p.Content
				break
			}
		}
	}
	return &CriticAgent{
		provider:     provider,
		systemPrompt: systemPrompt,
	}
}

// Review evalúa si los cambios y el contenido optimizado están respaldados semánticamente.
func (c *CriticAgent) Review(ctx context.Context, originalCV string, optResult *domain.OptimizedCVResult, gapAnswers []domain.GapAnswer) (domain.CriticVerdict, error) {
	changesJSON, _ := json.Marshal(optResult.Changes)
	answersText := FormatVerifiedContext(gapAnswers)

	userMsg := fmt.Sprintf(
		"CV ORIGINAL:\n%s\n\n%s\n\nCAMBIOS PROPUESTOS:\n%s\n\nCV OPTIMIZADO:\n%s",
		originalCV,
		answersText,
		string(changesJSON),
		optResult.ContentMarkdown,
	)

	raw, err := c.provider.CompleteJSON(ctx, c.systemPrompt, userMsg)
	if err != nil {
		return domain.CriticVerdict{}, fmt.Errorf("critic agent error: %w", err)
	}

	cleaned := ai.CleanJSONResponse(raw)
	var verdict domain.CriticVerdict
	if err := json.Unmarshal([]byte(cleaned), &verdict); err != nil {
		return domain.CriticVerdict{}, fmt.Errorf("error parseando veredicto de critic agent: %w", err)
	}

	return verdict, nil
}
