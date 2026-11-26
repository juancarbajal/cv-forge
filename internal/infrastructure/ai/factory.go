package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/jcarbajal/cv-checker/internal/domain"
)

type AIFactory struct {
	configRepo domain.ConfigRepository
	encryptor  domain.KeyEncryptor
}

func NewAIFactory(configRepo domain.ConfigRepository, encryptor domain.KeyEncryptor) *AIFactory {
	return &AIFactory{
		configRepo: configRepo,
		encryptor:  encryptor,
	}
}

// GetActiveProvider resuelve el proveedor de IA configurado como activo actualmente
func (f *AIFactory) GetActiveProvider(ctx context.Context) (domain.AIProvider, error) {
	activeConfig, err := f.configRepo.GetActiveConfig(ctx)
	if err != nil {
		// Si no hay configuración activa, retornar MockAIProvider por defecto para permitir uso sin configuración previa
		return NewMockAIProvider(), nil
	}

	provider := strings.ToLower(strings.TrimSpace(activeConfig.ProviderName))

	var apiKey string
	if len(activeConfig.EncryptedAPIKey) > 0 {
		decrypted, err := f.encryptor.Decrypt(activeConfig.EncryptedAPIKey)
		if err == nil {
			apiKey = decrypted
		}
	}

	switch provider {
	case "gemini":
		if apiKey == "" {
			return nil, fmt.Errorf("la API key de Gemini no está configurada")
		}
		return NewGeminiProvider(apiKey, activeConfig.DefaultModel), nil

	case "openai":
		if apiKey == "" {
			return nil, fmt.Errorf("la API key de OpenAI no está configurada")
		}
		baseURL := activeConfig.BaseURL
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		return NewOpenAICompatibleProvider("openai", baseURL, apiKey, activeConfig.DefaultModel), nil

	case "groq":
		if apiKey == "" {
			return nil, fmt.Errorf("la API key de Groq no está configurada")
		}
		baseURL := activeConfig.BaseURL
		if baseURL == "" {
			baseURL = "https://api.groq.com/openai/v1"
		}
		return NewOpenAICompatibleProvider("groq", baseURL, apiKey, activeConfig.DefaultModel), nil

	case "ollama":
		baseURL := activeConfig.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		return NewOllamaProvider(baseURL, activeConfig.DefaultModel), nil

	case "mock":
		return NewMockAIProvider(), nil

	default:
		return nil, domain.ErrUnsupportedProvider
	}
}
