package ai

import (
	"context"
	"testing"

	"github.com/jcarbajal/cv-checker/internal/domain"
)

type dummyConfigRepo struct {
	config *domain.ProviderConfig
}

func (d *dummyConfigRepo) SaveConfig(ctx context.Context, config *domain.ProviderConfig) error {
	d.config = config
	return nil
}
func (d *dummyConfigRepo) GetConfig(ctx context.Context, providerName string) (*domain.ProviderConfig, error) {
	return d.config, nil
}
func (d *dummyConfigRepo) ListConfigs(ctx context.Context) ([]domain.ProviderConfig, error) {
	return []domain.ProviderConfig{*d.config}, nil
}
func (d *dummyConfigRepo) SetActiveProvider(ctx context.Context, providerName string) error {
	return nil
}
func (d *dummyConfigRepo) GetActiveConfig(ctx context.Context) (*domain.ProviderConfig, error) {
	if d.config == nil {
		return nil, domain.ErrProviderNotActive
	}
	return d.config, nil
}

type dummyEncryptor struct{}

func (d *dummyEncryptor) Encrypt(p string) ([]byte, error) { return []byte(p), nil }
func (d *dummyEncryptor) Decrypt(c []byte) (string, error) { return string(c), nil }

func TestAIFactory(t *testing.T) {
	repo := &dummyConfigRepo{}
	factory := NewAIFactory(repo, &dummyEncryptor{})

	// 1. Sin config activa -> MockProvider
	prov, err := factory.GetActiveProvider(context.Background())
	if err != nil {
		t.Fatalf("error obteniendo mock default: %v", err)
	}
	if prov.ProviderName() != "mock" {
		t.Errorf("esperaba mock provider, obtuve %s", prov.ProviderName())
	}

	// 2. Configurada Ollama (sin API key requerida)
	repo.config = &domain.ProviderConfig{
		ProviderName: "ollama",
		DefaultModel: "llama3",
		IsActive:     true,
	}
	prov, err = factory.GetActiveProvider(context.Background())
	if err != nil {
		t.Fatalf("error obteniendo ollama provider: %v", err)
	}
	if prov.ProviderName() != "ollama" {
		t.Errorf("esperaba ollama provider, obtuve %s", prov.ProviderName())
	}
}

func TestCleanJSONResponse(t *testing.T) {
	raw := "```json\n{\"foo\": \"bar\"}\n```"
	clean := CleanJSONResponse(raw)
	if clean != "{\"foo\": \"bar\"}" {
		t.Errorf("CleanJSONResponse falló: '%s'", clean)
	}
}
