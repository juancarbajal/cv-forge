package app

import (
	"context"

	"github.com/jcarbajal/cv-checker/internal/domain"
)

type AuditService struct {
	repo domain.AuditRepository
}

func NewAuditService(repo domain.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Log(ctx context.Context, entry *domain.AuditLog) error {
	return s.repo.Log(ctx, entry)
}

func (s *AuditService) List(ctx context.Context, limit int) ([]domain.AuditLog, error) {
	return s.repo.List(ctx, limit)
}
