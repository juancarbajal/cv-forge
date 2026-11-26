package domain

import (
	"context"
	"time"
)

type ProviderConfig struct {
	ProviderName    string    `json:"provider_name"`
	EncryptedAPIKey []byte    `json:"encrypted_api_key"`
	DefaultModel    string    `json:"default_model"`
	BaseURL         string    `json:"base_url,omitempty"`
	IsActive        bool      `json:"is_active"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CVRepository interface {
	Create(ctx context.Context, cv *CV) error
	Update(ctx context.Context, cv *CV) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*CV, error)
	List(ctx context.Context, search string) ([]CV, error)
	AddVersion(ctx context.Context, version *CVVersion) error
	GetVersion(ctx context.Context, versionID string) (*CVVersion, error)
	ListVersions(ctx context.Context, cvID string) ([]CVVersion, error)
	SetFavoriteVersion(ctx context.Context, cvID string, versionID string) error
}

type CVFileStore interface {
	SaveContent(ctx context.Context, cvID string, version int, content string) (string, error)
	GetContent(ctx context.Context, filePath string) (string, error)
	DeleteContent(ctx context.Context, filePath string) error
}

type JobRepository interface {
	Create(ctx context.Context, job *JobApplication) error
	Update(ctx context.Context, job *JobApplication) error
	GetByID(ctx context.Context, id string) (*JobApplication, error)
	List(ctx context.Context) ([]JobApplication, error)
	SaveRequirements(ctx context.Context, jobAppID string, reqs []Requirement) error
	GetRequirements(ctx context.Context, jobAppID string) ([]Requirement, error)
	SaveMatchResults(ctx context.Context, jobAppID string, matches []MatchResult) error
	GetMatchResults(ctx context.Context, jobAppID string) ([]MatchResult, error)
	SaveATSScore(ctx context.Context, score *ATSScore) error
	GetATSScore(ctx context.Context, jobAppID string) (*ATSScore, error)
	SaveOptimizationChanges(ctx context.Context, changes []OptimizationChange) error
	GetOptimizationChanges(ctx context.Context, jobAppID string) ([]OptimizationChange, error)
}

type PromptRepository interface {
	Create(ctx context.Context, prompt *Prompt) error
	GetActive(ctx context.Context, pType PromptType) (*Prompt, error)
	GetByVersion(ctx context.Context, pType PromptType, version int) (*Prompt, error)
	ListByType(ctx context.Context, pType PromptType) ([]Prompt, error)
	SetActive(ctx context.Context, pType PromptType, version int) error
	SeedDefaults(ctx context.Context, defaultPrompts []Prompt) error
}

type AuditRepository interface {
	Log(ctx context.Context, entry *AuditLog) error
	List(ctx context.Context, limit int) ([]AuditLog, error)
}

type ConfigRepository interface {
	SaveConfig(ctx context.Context, config *ProviderConfig) error
	GetConfig(ctx context.Context, providerName string) (*ProviderConfig, error)
	ListConfigs(ctx context.Context) ([]ProviderConfig, error)
	SetActiveProvider(ctx context.Context, providerName string) error
	GetActiveConfig(ctx context.Context) (*ProviderConfig, error)
}

type KeyEncryptor interface {
	Encrypt(plainText string) ([]byte, error)
	Decrypt(cipherText []byte) (string, error)
}

type DiffEngine interface {
	GenerateDiff(original, modified string) (string, error)
}

type AIProvider interface {
	DetectLanguage(ctx context.Context, text string) (Language, error)
	ExtractAndClassifyRequirements(ctx context.Context, jobPosting string, prompt string) ([]Requirement, error)
	MatchCV(ctx context.Context, cvContent string, requirements []Requirement, prompt string) ([]MatchResult, error)
	GenerateRecommendations(ctx context.Context, cvContent string, matches []MatchResult, prompt string) (Recommendations, error)
	GenerateOptimizedCV(ctx context.Context, cvContent string, jobPosting string, matches []MatchResult, targetLang Language, prompt string) (OptimizedCVResult, error)
	CompleteJSON(ctx context.Context, systemPrompt, userMessage string) (string, error)
	ProviderName() string
}
