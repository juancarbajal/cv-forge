package agents

import (
	"context"
	"fmt"
	"strings"

	"github.com/jcarbajal/cv-checker/internal/domain"
)

// ReflectionResult agrupa el resultado final del bucle de optimización y su traza de agentes.
type ReflectionResult struct {
	Optimized     domain.OptimizedCVResult
	Alerts        []domain.ValidationAlert
	CriticVerdict domain.CriticVerdict
	Traces        []domain.AgentStepTrace
}

// ReflectionOptimizerAgent ejecuta un bucle acotado de generación, validación determinista y auditoría semántica.
type ReflectionOptimizerAgent struct {
	provider      domain.AIProvider
	validator     *domain.AntiInventionValidator
	critic        *CriticAgent
	maxIterations int
}

// NewReflectionOptimizerAgent crea un nuevo agente optimizador con reflexión.
func NewReflectionOptimizerAgent(
	provider domain.AIProvider,
	validator *domain.AntiInventionValidator,
	critic *CriticAgent,
	maxIterations int,
) *ReflectionOptimizerAgent {
	if maxIterations <= 0 {
		maxIterations = 2
	}
	return &ReflectionOptimizerAgent{
		provider:      provider,
		validator:     validator,
		critic:        critic,
		maxIterations: maxIterations,
	}
}

// Optimize ejecuta hasta maxIterations rondas de optimización y auditoría hasta lograr una versión aprobada.
func (a *ReflectionOptimizerAgent) Optimize(
	ctx context.Context,
	cvContent string,
	jobPosting string,
	matches []domain.MatchResult,
	targetLang domain.Language,
	optPrompt string,
	gapAnswers []domain.GapAnswer,
) (ReflectionResult, error) {
	var result ReflectionResult
	var feedback string
	var lastValidationErr error
	hasValidCandidate := false

	verifiedBlock := FormatVerifiedContext(gapAnswers)
	evidenceSource := cvContent
	if len(gapAnswers) > 0 {
		evidenceSource = cvContent + "\n\n" + verifiedBlock
	}

	for iter := 1; iter <= a.maxIterations; iter++ {
		effectivePrompt := optPrompt
		if len(gapAnswers) > 0 {
			effectivePrompt += "\n\n" + verifiedBlock
		}
		if feedback != "" {
			effectivePrompt += "\n\nRETROALIMENTACIÓN OBLIGATORIA DE AUDITORÍA PREVIA (CORREGIR EN ESTA VERSIÓN):\n" + feedback
		}

		optResult, err := a.provider.GenerateOptimizedCV(ctx, cvContent, jobPosting, matches, targetLang, effectivePrompt)
		if err != nil {
			result.Traces = append(result.Traces, domain.AgentStepTrace{
				Iteration: iter,
				AgentName: "ReflectionOptimizer",
				Status:    "FAILED",
				Notes:     err.Error(),
			})
			return result, err
		}
		result.Traces = append(result.Traces, domain.AgentStepTrace{
			Iteration: iter,
			AgentName: "ReflectionOptimizer",
			Status:    "COMPLETED",
			Notes:     fmt.Sprintf("Borrador generado con %d cambios propuestos", len(optResult.Changes)),
		})

		alerts, valErr := a.validator.Validate(evidenceSource, &optResult)
		if valErr != nil {
			lastValidationErr = valErr
			feedback = fmt.Sprintf("Error de validación anti-invención: %s", valErr.Error())
			result.Traces = append(result.Traces, domain.AgentStepTrace{
				Iteration: iter,
				AgentName: "AntiInventionValidator",
				Status:    "REVISE",
				Notes:     valErr.Error(),
			})
			continue
		}

		// Guardar al menos el último candidato que superó la validación determinista
		result.Optimized = optResult
		result.Alerts = alerts
		hasValidCandidate = true

		verdict, criticErr := a.critic.Review(ctx, cvContent, &optResult, gapAnswers)
		if criticErr != nil {
			// Si el critic falla por red/parseo, no bloqueamos el candidato determinista válido
			result.Traces = append(result.Traces, domain.AgentStepTrace{
				Iteration: iter,
				AgentName: "CriticAgent",
				Status:    "WARNING",
				Notes:     fmt.Sprintf("Auditoría semántica omitida por error: %v", criticErr),
			})
			break
		}

		result.CriticVerdict = verdict
		if verdict.Approved && !hasBlockingIssue(verdict.Issues) {
			result.Traces = append(result.Traces, domain.AgentStepTrace{
				Iteration: iter,
				AgentName: "CriticAgent",
				Status:    "APPROVED",
				Notes:     verdict.Summary,
			})
			break
		}

		feedback = formatCriticFeedback(verdict)
		result.Traces = append(result.Traces, domain.AgentStepTrace{
			Iteration: iter,
			AgentName: "CriticAgent",
			Status:    "REVISE",
			Notes:     feedback,
		})
	}

	if !hasValidCandidate {
		return result, lastValidationErr
	}
	return result, nil
}

func hasBlockingIssue(issues []domain.CriticIssue) bool {
	for _, issue := range issues {
		if strings.EqualFold(issue.Severity, "BLOCKING") {
			return true
		}
	}
	return false
}

func formatCriticFeedback(verdict domain.CriticVerdict) string {
	var parts []string
	if verdict.Summary != "" {
		parts = append(parts, verdict.Summary)
	}
	for _, iss := range verdict.Issues {
		parts = append(parts, fmt.Sprintf("[%s] %s -> %s (Sugerencia: %s)", iss.Severity, iss.Claim, iss.Reason, iss.Suggestion))
	}
	return strings.Join(parts, "; ")
}
