package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jcarbajal/cv-checker/internal/domain"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/ai"
)

type OptimizationPipelineService struct {
	jobRepo    domain.JobRepository
	cvService  *CVService
	promptRepo domain.PromptRepository
	aiFactory  *ai.AIFactory
	validator  *domain.AntiInventionValidator
	diffEngine domain.DiffEngine
	auditRepo  domain.AuditRepository
}

func NewOptimizationPipelineService(
	jobRepo domain.JobRepository,
	cvService *CVService,
	promptRepo domain.PromptRepository,
	aiFactory *ai.AIFactory,
	diffEngine domain.DiffEngine,
	auditRepo domain.AuditRepository,
) *OptimizationPipelineService {
	return &OptimizationPipelineService{
		jobRepo:    jobRepo,
		cvService:  cvService,
		promptRepo: promptRepo,
		aiFactory:  aiFactory,
		validator:  domain.NewAntiInventionValidator(),
		diffEngine: diffEngine,
		auditRepo:  auditRepo,
	}
}

// AnalyzeJobOffer ejecuta la Fase 1: Detección de idioma y extracción de requisitos
func (p *OptimizationPipelineService) AnalyzeJobOffer(ctx context.Context, jobID string) ([]domain.Requirement, domain.Language, error) {
	job, err := p.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, "", err
	}

	provider, err := p.aiFactory.GetActiveProvider(ctx)
	if err != nil {
		return nil, "", err
	}

	promptObj, err := p.promptRepo.GetActive(ctx, domain.PromptTypeJobAnalysis)
	promptContent := ""
	if err == nil && promptObj != nil {
		promptContent = promptObj.Content
	}

	// 1. Detección de idioma (RF-009)
	lang, err := provider.DetectLanguage(ctx, job.JobPostingRaw)
	if err != nil {
		lang = domain.LangSpanish
	}
	job.DetectedLanguage = lang
	if job.TargetLanguage == "" {
		job.TargetLanguage = lang
	}

	// 2. Extracción y clasificación de requisitos (RF-010, RF-011)
	reqs, err := provider.ExtractAndClassifyRequirements(ctx, job.JobPostingRaw, promptContent)
	if err != nil {
		p.auditRepo.Log(ctx, &domain.AuditLog{
			Action:       domain.ActionAnalyzeJob,
			Provider:     provider.ProviderName(),
			ResultStatus: "FAILED",
			DetailsJSON:  fmt.Sprintf(`{"error": "%s"}`, err.Error()),
		})
		return nil, lang, fmt.Errorf("error extrayendo requisitos: %w", err)
	}

	for i := range reqs {
		reqs[i].JobApplicationID = job.ID
	}

	if err := p.jobRepo.SaveRequirements(ctx, job.ID, reqs); err != nil {
		return nil, lang, err
	}

	job.Status = domain.StatusAnalyzed
	if err := p.jobRepo.Update(ctx, job); err != nil {
		return nil, lang, err
	}

	p.auditRepo.Log(ctx, &domain.AuditLog{
		Action:       domain.ActionAnalyzeJob,
		Provider:     provider.ProviderName(),
		ResultStatus: "SUCCESS",
		DetailsJSON:  fmt.Sprintf(`{"requirements_extracted": %d, "language": "%s"}`, len(reqs), lang),
	})

	return reqs, lang, nil
}

// MatchCVAndScore ejecuta la Fase 2 y 3: Matching semántico, cálculo ATS y recomendaciones
func (p *OptimizationPipelineService) MatchCVAndScore(ctx context.Context, jobID, versionID string) (*domain.ATSScore, []domain.MatchResult, *domain.Recommendations, error) {
	job, err := p.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, nil, nil, err
	}

	cvContent, _, err := p.cvService.GetVersionContent(ctx, versionID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error obteniendo contenido del CV: %w", err)
	}

	reqs, err := p.jobRepo.GetRequirements(ctx, jobID)
	if err != nil || len(reqs) == 0 {
		return nil, nil, nil, fmt.Errorf("la postulación no tiene requisitos analizados")
	}

	provider, err := p.aiFactory.GetActiveProvider(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	matchPromptObj, _ := p.promptRepo.GetActive(ctx, domain.PromptTypeCVMatching)
	matchPrompt := ""
	if matchPromptObj != nil {
		matchPrompt = matchPromptObj.Content
	}

	// 1. Matching (RF-012, RF-013, RF-014)
	matches, err := provider.MatchCV(ctx, cvContent, reqs, matchPrompt)
	if err != nil {
		p.auditRepo.Log(ctx, &domain.AuditLog{
			Action:       domain.ActionMatchCV,
			Provider:     provider.ProviderName(),
			ResultStatus: "FAILED",
			DetailsJSON:  fmt.Sprintf(`{"error": "%s"}`, err.Error()),
		})
		return nil, nil, nil, fmt.Errorf("error ejecutando matching: %w", err)
	}

	for i := range matches {
		matches[i].JobApplicationID = jobID
	}

	if err := p.jobRepo.SaveMatchResults(ctx, jobID, matches); err != nil {
		return nil, nil, nil, err
	}

	// 2. Cálculo ATS Score determinista con pesos fijos (RF-015, RF-016, RF-017)
	atsScore := domain.CalculateATSScore(jobID, reqs, matches)

	// 3. Recomendaciones y Keywords (RF-018)
	recPromptObj, _ := p.promptRepo.GetActive(ctx, domain.PromptTypeATSScoring)
	recPrompt := ""
	if recPromptObj != nil {
		recPrompt = recPromptObj.Content
	}
	recs, err := provider.GenerateRecommendations(ctx, cvContent, matches, recPrompt)
	if err == nil {
		recsBytes, _ := json.Marshal(recs)
		atsScore.RecommendationsJSON = string(recsBytes)
	}

	if err := p.jobRepo.SaveATSScore(ctx, &atsScore); err != nil {
		return nil, nil, nil, err
	}

	job.TotalATSScore = atsScore.TotalScore
	job.CVVersionID = versionID
	if err := p.jobRepo.Update(ctx, job); err != nil {
		return nil, nil, nil, err
	}

	p.auditRepo.Log(ctx, &domain.AuditLog{
		Action:       domain.ActionMatchCV,
		Provider:     provider.ProviderName(),
		ResultStatus: "SUCCESS",
		DetailsJSON:  fmt.Sprintf(`{"total_ats_score": %.2f}`, atsScore.TotalScore),
	})

	return &atsScore, matches, &recs, nil
}

type OptimizationExecutionResult struct {
	NewVersion     *domain.CVVersion           `json:"new_version"`
	DiffUnified    string                      `json:"diff_unified"`
	Changes        []domain.OptimizationChange `json:"changes"`
	ValidationWarn []domain.ValidationAlert    `json:"validation_warnings"`
}

// OptimizeAndGenerateCV ejecuta la Fase 4 y 5: Optimización con regla Anti-Invención, Diff y versionado
func (p *OptimizationPipelineService) OptimizeAndGenerateCV(ctx context.Context, jobID, versionID string, targetLang domain.Language) (*OptimizationExecutionResult, error) {
	job, err := p.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	cvContent, sourceVer, err := p.cvService.GetVersionContent(ctx, versionID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo contenido del CV origen: %w", err)
	}

	matches, err := p.jobRepo.GetMatchResults(ctx, jobID)
	if err != nil || len(matches) == 0 {
		return nil, fmt.Errorf("debe ejecutar el matching previo antes de optimizar")
	}

	provider, err := p.aiFactory.GetActiveProvider(ctx)
	if err != nil {
		return nil, err
	}

	optPromptObj, _ := p.promptRepo.GetActive(ctx, domain.PromptTypeCVOptimization)
	optPrompt := ""
	if optPromptObj != nil {
		optPrompt = optPromptObj.Content
	}

	if targetLang == "" {
		targetLang = job.TargetLanguage
	}
	if targetLang == "" {
		targetLang = job.DetectedLanguage
	}

	// 1. Generación de CV Optimizado (RF-020, RF-021, RF-022)
	optResult, err := provider.GenerateOptimizedCV(ctx, cvContent, job.JobPostingRaw, matches, targetLang, optPrompt)
	if err != nil {
		p.auditRepo.Log(ctx, &domain.AuditLog{
			Action:       domain.ActionOptimizeCV,
			Provider:     provider.ProviderName(),
			ResultStatus: "FAILED",
			DetailsJSON:  fmt.Sprintf(`{"error": "%s"}`, err.Error()),
		})
		return nil, fmt.Errorf("error generando CV optimizado: %w", err)
	}

	// 2. Validación de regla Anti-Invención (RF-019)
	alerts, err := p.validator.Validate(cvContent, &optResult)
	if err != nil {
		p.auditRepo.Log(ctx, &domain.AuditLog{
			Action:       domain.ActionOptimizeCV,
			Provider:     provider.ProviderName(),
			ResultStatus: "FAILED",
			DetailsJSON:  fmt.Sprintf(`{"anti_invention_error": "%s"}`, err.Error()),
		})
		return nil, fmt.Errorf("validación anti-invención no superada: %w", err)
	}

	// 3. Generación de Diff estilo Git (RF-023)
	diffStr, err := p.diffEngine.GenerateDiff(cvContent, optResult.ContentMarkdown)
	if err != nil {
		diffStr = "Diff no disponible"
	}

	// 4. Guardar nueva versión de CV (RF-003)
	reason := fmt.Sprintf("Optimizado para %s (%s)", job.Company, job.Role)
	newVer, err := p.cvService.AddNewVersion(ctx, sourceVer.CVID, sourceVer.ID, optResult.ContentMarkdown, reason, provider.ProviderName(), "active-model")
	if err != nil {
		return nil, fmt.Errorf("error guardando nueva versión: %w", err)
	}

	// 5. Guardar trazabilidad de cambios (RF-024)
	for i := range optResult.Changes {
		optResult.Changes[i].JobApplicationID = jobID
		optResult.Changes[i].CVVersionGeneratedID = newVer.ID
	}
	_ = p.jobRepo.SaveOptimizationChanges(ctx, optResult.Changes)

	job.Status = domain.StatusOptimized
	job.CVVersionID = newVer.ID
	job.UpdatedAt = time.Now().UTC()
	_ = p.jobRepo.Update(ctx, job)

	p.auditRepo.Log(ctx, &domain.AuditLog{
		Action:       domain.ActionOptimizeCV,
		Provider:     provider.ProviderName(),
		ResultStatus: "SUCCESS",
		DetailsJSON:  fmt.Sprintf(`{"new_version_id": "%s", "changes_recorded": %d}`, newVer.ID, len(optResult.Changes)),
	})

	return &OptimizationExecutionResult{
		NewVersion:     newVer,
		DiffUnified:    diffStr,
		Changes:        optResult.Changes,
		ValidationWarn: alerts,
	}, nil
}
