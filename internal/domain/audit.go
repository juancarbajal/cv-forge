package domain

import (
	"time"
)

type AuditAction string

const (
	ActionAnalyzeJob   AuditAction = "ANALYZE_JOB"
	ActionMatchCV      AuditAction = "MATCH_CV"
	ActionOptimizeCV   AuditAction = "OPTIMIZE_CV"
	ActionTranslateCV  AuditAction = "TRANSLATE_CV"
	ActionCreateCV     AuditAction = "CREATE_CV"
	ActionUpdateCV     AuditAction = "UPDATE_CV"
	ActionUpdatePrompt AuditAction = "UPDATE_PROMPT"
	ActionUpdateConfig AuditAction = "UPDATE_CONFIG"
)

type AuditLog struct {
	ID           string      `json:"id"`
	Timestamp    time.Time   `json:"timestamp"`
	Action       AuditAction `json:"action"`
	Provider     string      `json:"provider,omitempty"`
	Model        string      `json:"model,omitempty"`
	ResultStatus string      `json:"result_status"` // "SUCCESS", "FAILED", "CANCELLED"
	DetailsJSON  string      `json:"details_json,omitempty"`
}
