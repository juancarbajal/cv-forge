package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcarbajal/cv-checker/internal/domain"
)

type CVService struct {
	repo      domain.CVRepository
	fileStore domain.CVFileStore
}

func NewCVService(repo domain.CVRepository, fileStore domain.CVFileStore) *CVService {
	return &CVService{
		repo:      repo,
		fileStore: fileStore,
	}
}

// CreateCV crea un nuevo CV con su versión inicial (v1) y guarda el contenido markdown en disco
func (s *CVService) CreateCV(ctx context.Context, title, description string, lang domain.Language, initialMarkdown string) (*domain.CV, *domain.CVVersion, error) {
	cvID := uuid.NewString()
	versionID := uuid.NewString()

	filePath, err := s.fileStore.SaveContent(ctx, cvID, 1, initialMarkdown)
	if err != nil {
		return nil, nil, fmt.Errorf("error guardando archivo inicial de CV: %w", err)
	}

	cv := &domain.CV{
		ID:                cvID,
		Title:             title,
		Description:       description,
		PrimaryLanguage:   lang,
		FavoriteVersionID: versionID,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := cv.Validate(); err != nil {
		s.fileStore.DeleteContent(ctx, filePath)
		return nil, nil, err
	}

	if err := s.repo.Create(ctx, cv); err != nil {
		s.fileStore.DeleteContent(ctx, filePath)
		return nil, nil, err
	}

	v1 := &domain.CVVersion{
		ID:            versionID,
		CVID:          cvID,
		VersionNumber: 1,
		FilePath:      filePath,
		ChangeReason:  "Creación inicial",
		IsFavorite:    true,
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.repo.AddVersion(ctx, v1); err != nil {
		return nil, nil, err
	}

	cv.Versions = []domain.CVVersion{*v1}
	return cv, v1, nil
}

// AddNewVersion crea una nueva versión de un CV existente a partir de nuevo contenido Markdown
func (s *CVService) AddNewVersion(ctx context.Context, cvID string, sourceVersionID string, markdownContent, changeReason, provider, model string) (*domain.CVVersion, error) {
	existingVersions, err := s.repo.ListVersions(ctx, cvID)
	if err != nil {
		return nil, err
	}

	nextVersionNum := len(existingVersions) + 1
	versionID := uuid.NewString()

	filePath, err := s.fileStore.SaveContent(ctx, cvID, nextVersionNum, markdownContent)
	if err != nil {
		return nil, fmt.Errorf("error guardando contenido de versión: %w", err)
	}

	version := &domain.CVVersion{
		ID:              versionID,
		CVID:            cvID,
		VersionNumber:   nextVersionNum,
		FilePath:        filePath,
		SourceVersionID: sourceVersionID,
		ChangeReason:    changeReason,
		AIProvider:      provider,
		AIModel:         model,
		IsFavorite:      false,
		CreatedAt:       time.Now().UTC(),
	}

	if err := s.repo.AddVersion(ctx, version); err != nil {
		s.fileStore.DeleteContent(ctx, filePath)
		return nil, err
	}

	return version, nil
}

func (s *CVService) DuplicateCV(ctx context.Context, cvID string, newTitle string) (*domain.CV, error) {
	origCV, err := s.repo.GetByID(ctx, cvID)
	if err != nil {
		return nil, err
	}

	// Obtener la versión favorita o la última versión
	var sourceVersion *domain.CVVersion
	if origCV.FavoriteVersionID != "" {
		sourceVersion, _ = s.repo.GetVersion(ctx, origCV.FavoriteVersionID)
	}
	if sourceVersion == nil && len(origCV.Versions) > 0 {
		sourceVersion = &origCV.Versions[len(origCV.Versions)-1]
	}

	content := ""
	if sourceVersion != nil {
		c, err := s.fileStore.GetContent(ctx, sourceVersion.FilePath)
		if err == nil {
			content = c
		}
	}

	newCV, _, err := s.CreateCV(ctx, newTitle, fmt.Sprintf("Copia de %s", origCV.Title), origCV.PrimaryLanguage, content)
	return newCV, err
}

func (s *CVService) SetFavoriteVersion(ctx context.Context, cvID, versionID string) error {
	return s.repo.SetFavoriteVersion(ctx, cvID, versionID)
}

func (s *CVService) GetCV(ctx context.Context, cvID string) (*domain.CV, error) {
	return s.repo.GetByID(ctx, cvID)
}

func (s *CVService) ListCVs(ctx context.Context, search string) ([]domain.CV, error) {
	return s.repo.List(ctx, search)
}

func (s *CVService) GetVersionContent(ctx context.Context, versionID string) (string, *domain.CVVersion, error) {
	v, err := s.repo.GetVersion(ctx, versionID)
	if err != nil {
		return "", nil, err
	}
	content, err := s.fileStore.GetContent(ctx, v.FilePath)
	if err != nil {
		return "", nil, err
	}
	return content, v, nil
}

func (s *CVService) DeleteCV(ctx context.Context, cvID string) error {
	versions, err := s.repo.ListVersions(ctx, cvID)
	if err == nil {
		for _, v := range versions {
			s.fileStore.DeleteContent(ctx, v.FilePath)
		}
	}
	return s.repo.Delete(ctx, cvID)
}
