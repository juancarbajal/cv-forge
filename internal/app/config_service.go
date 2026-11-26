package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jcarbajal/cv-checker/internal/domain"
)

type ConfigSecurityService struct {
	repo      domain.ConfigRepository
	encryptor domain.KeyEncryptor
	auditRepo domain.AuditRepository
}

func NewConfigSecurityService(repo domain.ConfigRepository, encryptor domain.KeyEncryptor, auditRepo domain.AuditRepository) *ConfigSecurityService {
	return &ConfigSecurityService{
		repo:      repo,
		encryptor: encryptor,
		auditRepo: auditRepo,
	}
}

// SaveProviderConfig guarda o actualiza la configuración de un proveedor cifrando su clave API
func (s *ConfigSecurityService) SaveProviderConfig(ctx context.Context, providerName, apiKey, defaultModel, baseURL string, isActive bool) error {
	var encryptedKey []byte
	var err error

	if apiKey != "" {
		encryptedKey, err = s.encryptor.Encrypt(apiKey)
		if err != nil {
			return fmt.Errorf("error cifrando API Key: %w", err)
		}
	}

	config := &domain.ProviderConfig{
		ProviderName:    providerName,
		EncryptedAPIKey: encryptedKey,
		DefaultModel:    defaultModel,
		BaseURL:         baseURL,
		IsActive:        isActive,
		UpdatedAt:       time.Now().UTC(),
	}

	if err := s.repo.SaveConfig(ctx, config); err != nil {
		return err
	}

	if isActive {
		_ = s.repo.SetActiveProvider(ctx, providerName)
	}

	_ = s.auditRepo.Log(ctx, &domain.AuditLog{
		Action:       domain.ActionUpdateConfig,
		Provider:     providerName,
		Model:        defaultModel,
		ResultStatus: "SUCCESS",
		DetailsJSON:  fmt.Sprintf(`{"is_active": %t}`, isActive),
	})

	return nil
}

func (s *ConfigSecurityService) SetActiveProvider(ctx context.Context, providerName string) error {
	err := s.repo.SetActiveProvider(ctx, providerName)
	if err == nil {
		_ = s.auditRepo.Log(ctx, &domain.AuditLog{
			Action:       domain.ActionUpdateConfig,
			Provider:     providerName,
			ResultStatus: "SUCCESS",
			DetailsJSON:  `{"active_switched": true}`,
		})
	}
	return err
}

func (s *ConfigSecurityService) ListConfigs(ctx context.Context) ([]domain.ProviderConfig, error) {
	return s.repo.ListConfigs(ctx)
}

func (s *ConfigSecurityService) GetActiveConfig(ctx context.Context) (*domain.ProviderConfig, error) {
	return s.repo.GetActiveConfig(ctx)
}
