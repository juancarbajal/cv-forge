package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jcarbajal/cv-checker/internal/domain"
)

type JobService struct {
	repo domain.JobRepository
}

func NewJobService(repo domain.JobRepository) *JobService {
	return &JobService{repo: repo}
}

func (s *JobService) CreateJob(ctx context.Context, company, role, url, jobPostingRaw string, cvID, cvVersionID string, targetLang domain.Language) (*domain.JobApplication, error) {
	jobID := uuid.NewString()

	job := &domain.JobApplication{
		ID:               jobID,
		Company:          company,
		Role:             role,
		URL:              url,
		JobPostingRaw:    jobPostingRaw,
		DetectedLanguage: domain.LangSpanish,
		TargetLanguage:   targetLang,
		CVID:             cvID,
		CVVersionID:      cvVersionID,
		Status:           domain.StatusDraft,
		TotalATSScore:    0.0,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	if err := job.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}

func (s *JobService) UpdateJob(ctx context.Context, job *domain.JobApplication) error {
	if err := job.Validate(); err != nil {
		return err
	}
	return s.repo.Update(ctx, job)
}

func (s *JobService) GetJob(ctx context.Context, jobID string) (*domain.JobApplication, error) {
	return s.repo.GetByID(ctx, jobID)
}

func (s *JobService) ListJobs(ctx context.Context) ([]domain.JobApplication, error) {
	return s.repo.List(ctx)
}

func (s *JobService) GetJobHistory(ctx context.Context, jobID string) (*domain.JobApplication, []domain.Requirement, []domain.MatchResult, *domain.ATSScore, []domain.OptimizationChange, error) {
	job, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	reqs, _ := s.repo.GetRequirements(ctx, jobID)
	matches, _ := s.repo.GetMatchResults(ctx, jobID)
	score, _ := s.repo.GetATSScore(ctx, jobID)
	changes, _ := s.repo.GetOptimizationChanges(ctx, jobID)

	return job, reqs, matches, score, changes, nil
}
