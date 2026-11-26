package domain

import (
	"strings"
	"time"
)

type PromptType string

const (
	PromptTypeJobAnalysis    PromptType = "JOB_ANALYSIS"
	PromptTypeCVMatching     PromptType = "CV_MATCHING"
	PromptTypeATSScoring     PromptType = "ATS_SCORING"
	PromptTypeTranslation    PromptType = "TRANSLATION"
	PromptTypeCVOptimization PromptType = "CV_OPTIMIZATION"
	PromptTypeCriticAudit    PromptType = "CRITIC_AUDIT"
	PromptTypeGapInterview   PromptType = "GAP_INTERVIEW"
)

type Prompt struct {
	ID        string     `json:"id"`
	Type      PromptType `json:"type"`
	Name      string     `json:"name"`
	Version   int        `json:"version"`
	Content   string     `json:"content"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
}

func (p *Prompt) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(string(p.Type)) == "" {
		return ErrInvalidInput
	}
	if p.Version <= 0 {
		return ErrInvalidInput
	}
	if strings.TrimSpace(p.Content) == "" {
		return ErrInvalidInput
	}
	return nil
}
