package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcarbajal/cv-checker/internal/domain"
)

type PromptService struct {
	repo       domain.PromptRepository
	diffEngine domain.DiffEngine
	auditRepo  domain.AuditRepository
}

func NewPromptService(repo domain.PromptRepository, diffEngine domain.DiffEngine, auditRepo domain.AuditRepository) *PromptService {
	return &PromptService{
		repo:       repo,
		diffEngine: diffEngine,
		auditRepo:  auditRepo,
	}
}

// GetActivePrompt obtiene la versión activa actual para un tipo de prompt
func (s *PromptService) GetActivePrompt(ctx context.Context, pType domain.PromptType) (*domain.Prompt, error) {
	return s.repo.GetActive(ctx, pType)
}

// CreateNewVersion crea una nueva versión de un prompt incrementando el número de versión
func (s *PromptService) CreateNewVersion(ctx context.Context, pType domain.PromptType, name, content string, makeActive bool) (*domain.Prompt, error) {
	existing, err := s.repo.ListByType(ctx, pType)
	if err != nil {
		return nil, err
	}

	nextVersion := 1
	for _, p := range existing {
		if p.Version >= nextVersion {
			nextVersion = p.Version + 1
		}
	}

	newPrompt := &domain.Prompt{
		ID:        uuid.NewString(),
		Type:      pType,
		Name:      name,
		Version:   nextVersion,
		Content:   content,
		IsActive:  makeActive,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, newPrompt); err != nil {
		return nil, err
	}

	_ = s.auditRepo.Log(ctx, &domain.AuditLog{
		Action:       domain.ActionUpdatePrompt,
		ResultStatus: "SUCCESS",
		DetailsJSON:  fmt.Sprintf(`{"prompt_type": "%s", "version": %d}`, pType, nextVersion),
	})

	return newPrompt, nil
}

// ListVersions lista el historial de versiones para un tipo de prompt
func (s *PromptService) ListVersions(ctx context.Context, pType domain.PromptType) ([]domain.Prompt, error) {
	return s.repo.ListByType(ctx, pType)
}

// RollbackToVersion activa una versión anterior de un tipo de prompt (RF-029)
func (s *PromptService) RollbackToVersion(ctx context.Context, pType domain.PromptType, version int) error {
	if err := s.repo.SetActive(ctx, pType, version); err != nil {
		return err
	}

	_ = s.auditRepo.Log(ctx, &domain.AuditLog{
		Action:       domain.ActionUpdatePrompt,
		ResultStatus: "SUCCESS",
		DetailsJSON:  fmt.Sprintf(`{"rollback_type": "%s", "active_version": %d}`, pType, version),
	})

	return nil
}

// CompareVersions genera un diff unificado estilo Git entre dos versiones del mismo prompt (RF-029)
func (s *PromptService) CompareVersions(ctx context.Context, pType domain.PromptType, v1, v2 int) (string, error) {
	p1, err := s.repo.GetByVersion(ctx, pType, v1)
	if err != nil {
		return "", fmt.Errorf("versión %d no encontrada: %w", v1, err)
	}
	p2, err := s.repo.GetByVersion(ctx, pType, v2)
	if err != nil {
		return "", fmt.Errorf("versión %d no encontrada: %w", v2, err)
	}

	return s.diffEngine.GenerateDiff(p1.Content, p2.Content)
}
