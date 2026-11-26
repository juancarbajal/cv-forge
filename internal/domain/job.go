package domain

import (
	"strings"
	"time"
)

type RequirementCategory string

const (
	CategoryHardSkill   RequirementCategory = "hard_skill"
	CategoryExperience  RequirementCategory = "experience"
	CategoryKeywords    RequirementCategory = "keywords"
	CategoryEducation   RequirementCategory = "education"
	CategoryLanguage    RequirementCategory = "language"
)

type RequirementPriority string

const (
	PriorityMandatory RequirementPriority = "obligatorio"
	PriorityDesirable RequirementPriority = "deseable"
	PriorityOptional  RequirementPriority = "opcional"
)

type Requirement struct {
	ID               string              `json:"id"`
	JobApplicationID string              `json:"job_application_id"`
	Name             string              `json:"name"`
	Category         RequirementCategory `json:"categoria"`
	Priority         RequirementPriority `json:"prioridad"`
}

type JobStatus string

const (
	StatusDraft     JobStatus = "Borrador"
	StatusAnalyzed  JobStatus = "Analizado"
	StatusOptimized JobStatus = "Optimizado"
	StatusApplied   JobStatus = "Postulado"
)

type JobApplication struct {
	ID               string        `json:"id"`
	Company          string        `json:"company"`
	Role             string        `json:"role"`
	URL              string        `json:"url,omitempty"`
	JobPostingRaw    string        `json:"job_posting_raw"`
	DetectedLanguage Language      `json:"detected_language"`
	TargetLanguage   Language      `json:"target_language"`
	CVID             string        `json:"cv_id"`
	CVVersionID      string        `json:"cv_version_id"`
	Status           JobStatus     `json:"status"`
	TotalATSScore    float64       `json:"total_ats_score"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
	Requirements     []Requirement `json:"requirements,omitempty"`
}

func (j *JobApplication) Validate() error {
	if strings.TrimSpace(j.ID) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(j.Company) == "" || strings.TrimSpace(j.Role) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(j.JobPostingRaw) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(j.CVID) == "" {
		return ErrInvalidInput
	}
	return nil
}
