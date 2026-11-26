package agents

import (
	"context"
	"strings"
	"testing"

	"github.com/jcarbajal/cv-checker/internal/domain"
)

type stubAgentProvider struct {
	optCalls    int
	criticCalls int
	jsonRespFn  func(sys, user string) (string, error)
}

func (s *stubAgentProvider) ProviderName() string { return "stub" }

func (s *stubAgentProvider) DetectLanguage(ctx context.Context, text string) (domain.Language, error) {
	return domain.LangSpanish, nil
}
func (s *stubAgentProvider) ExtractAndClassifyRequirements(ctx context.Context, jobPosting string, prompt string) ([]domain.Requirement, error) {
	return nil, nil
}
func (s *stubAgentProvider) MatchCV(ctx context.Context, cvContent string, requirements []domain.Requirement, prompt string) ([]domain.MatchResult, error) {
	return nil, nil
}
func (s *stubAgentProvider) GenerateRecommendations(ctx context.Context, cvContent string, matches []domain.MatchResult, prompt string) (domain.Recommendations, error) {
	return domain.Recommendations{}, nil
}
func (s *stubAgentProvider) GenerateOptimizedCV(ctx context.Context, cvContent string, jobPosting string, matches []domain.MatchResult, targetLang domain.Language, prompt string) (domain.OptimizedCVResult, error) {
	s.optCalls++
	return domain.OptimizedCVResult{
		ContentMarkdown: "# CV Optimizado\nExperiencia en Go y PostgreSQL.",
		TargetLanguage:  targetLang,
		Changes: []domain.OptimizationChange{
			{
				ChangeSummary: "Destacar Go",
				EvidenceUsed:  "Experiencia en Go",
			},
		},
	}, nil
}
func (s *stubAgentProvider) CompleteJSON(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	if s.jsonRespFn != nil {
		return s.jsonRespFn(systemPrompt, userMessage)
	}
	s.criticCalls++
	return `{"approved": true, "issues": [], "summary": "OK"}`, nil
}

func TestReflectionOptimizerAgent_RetriesAndApproves(t *testing.T) {
	callCount := 0
	prov := &stubAgentProvider{
		jsonRespFn: func(sys, user string) (string, error) {
			callCount++
			if callCount == 1 {
				return `{"approved": false, "issues": [{"severity": "BLOCKING", "claim": "5 años", "reason": "No consta", "suggestion": "Quitar 5 años"}], "summary": "Revisar seniority"}`, nil
			}
			return `{"approved": true, "issues": [], "summary": "Corregido y verificado"}`, nil
		},
	}

	critic := NewCriticAgent(prov, "")
	opt := NewReflectionOptimizerAgent(prov, domain.NewAntiInventionValidator(), critic, 2)

	res, err := opt.Optimize(context.Background(), "Experiencia en Go", "Buscamos Go dev", nil, domain.LangSpanish, "prompt", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prov.optCalls != 2 {
		t.Fatalf("expected 2 optimization calls, got %d", prov.optCalls)
	}
	if !res.CriticVerdict.Approved {
		t.Fatalf("expected final verdict to be approved")
	}
	if len(res.Traces) != 4 {
		t.Fatalf("expected 4 trace entries (2 per iteration), got %d", len(res.Traces))
	}
}

func TestGapInterviewerAgent_SkipsWhenNoGaps(t *testing.T) {
	prov := &stubAgentProvider{
		jsonRespFn: func(sys, user string) (string, error) {
			t.Fatalf("should not call provider when there are no gaps")
			return "", nil
		},
	}
	interviewer := NewGapInterviewerAgent(prov, "")
	questions, err := interviewer.GenerateQuestions(context.Background(), "CV", []domain.MatchResult{
		{RequirementName: "Go", Status: domain.MatchStatusFulfilled},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(questions) != 0 {
		t.Fatalf("expected 0 questions, got %d", len(questions))
	}
}

func TestGapInterviewerAgent_GeneratesQuestionsForGaps(t *testing.T) {
	prov := &stubAgentProvider{
		jsonRespFn: func(sys, user string) (string, error) {
			if !strings.Contains(user, "Docker") {
				t.Fatalf("expected Docker gap in user message")
			}
			return `{"questions": [{"requirement_name": "Docker", "question": "¿Usaste Docker?", "why_asked": "Falta evidencia"}]}`, nil
		},
	}
	interviewer := NewGapInterviewerAgent(prov, "")
	questions, err := interviewer.GenerateQuestions(context.Background(), "CV", []domain.MatchResult{
		{RequirementName: "Go", Status: domain.MatchStatusFulfilled},
		{RequirementName: "Docker", Status: domain.MatchStatusMissing},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(questions) != 1 || questions[0].RequirementName != "Docker" {
		t.Fatalf("unexpected questions: %+v", questions)
	}
}
