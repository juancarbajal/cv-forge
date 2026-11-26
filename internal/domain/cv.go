package domain

import (
	"strings"
	"time"
)

// Language representa el código de idioma (es, en, pt)
type Language string

const (
	LangSpanish    Language = "es"
	LangEnglish    Language = "en"
	LangPortuguese Language = "pt"
	LangUnknown    Language = "unknown"
)

// CV representa la entidad raíz de un currículum
type CV struct {
	ID                string      `json:"id"`
	Title             string      `json:"title"`
	Description       string      `json:"description"`
	PrimaryLanguage   Language    `json:"primary_language"`
	FavoriteVersionID string      `json:"favorite_version_id,omitempty"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	Versions          []CVVersion `json:"versions,omitempty"`
}

// CVVersion representa una versión histórica u optimizada de un CV
type CVVersion struct {
	ID              string    `json:"id"`
	CVID            string    `json:"cv_id"`
	VersionNumber   int       `json:"version_number"`
	FilePath        string    `json:"file_path"`
	SourceVersionID string    `json:"source_version_id,omitempty"`
	ChangeReason    string    `json:"change_reason"`
	AIProvider      string    `json:"ai_provider,omitempty"`
	AIModel         string    `json:"ai_model,omitempty"`
	IsFavorite      bool      `json:"is_favorite"`
	CreatedAt       time.Time `json:"created_at"`
}

func (c *CV) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(c.Title) == "" {
		return ErrInvalidInput
	}
	if c.PrimaryLanguage == "" {
		c.PrimaryLanguage = LangSpanish
	}
	return nil
}

func (v *CVVersion) Validate() error {
	if strings.TrimSpace(v.ID) == "" || strings.TrimSpace(v.CVID) == "" {
		return ErrInvalidInput
	}
	if v.VersionNumber <= 0 {
		return ErrInvalidInput
	}
	if strings.TrimSpace(v.FilePath) == "" {
		return ErrInvalidInput
	}
	return nil
}
