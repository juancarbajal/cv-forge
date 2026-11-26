package sqlite

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcarbajal/cv-checker/internal/domain"
)

func setupTestDB(t *testing.T) (*SQLiteCVRepository, *SQLiteJobRepository, *SQLitePromptRepository, *SQLiteAuditRepository, *SQLiteConfigRepository, func()) {
	t.Helper()
	tempFile, err := os.CreateTemp("", "test_cv_checker_*.db")
	if err != nil {
		t.Fatalf("error creando temp db: %v", err)
	}
	dbPath := tempFile.Name()
	tempFile.Close()

	db, err := OpenDB(dbPath)
	if err != nil {
		os.Remove(dbPath)
		t.Fatalf("error abriendo DB: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	cvRepo := NewSQLiteCVRepository(db)
	jobRepo := NewSQLiteJobRepository(db)
	promptRepo := NewSQLitePromptRepository(db)
	auditRepo := NewSQLiteAuditRepository(db)
	configRepo := NewSQLiteConfigRepository(db)

	return cvRepo, jobRepo, promptRepo, auditRepo, configRepo, cleanup
}

func TestSQLiteRepositories(t *testing.T) {
	cvRepo, jobRepo, promptRepo, auditRepo, configRepo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Test CV & Versions
	cvID := uuid.NewString()
	cv := &domain.CV{
		ID:              cvID,
		Title:           "CV Backend Go",
		Description:     "Perfil especializado en backend",
		PrimaryLanguage: domain.LangSpanish,
	}
	if err := cvRepo.Create(ctx, cv); err != nil {
		t.Fatalf("error creando CV: %v", err)
	}

	v1ID := uuid.NewString()
	ver1 := &domain.CVVersion{
		ID:            v1ID,
		CVID:          cvID,
		VersionNumber: 1,
		FilePath:      "data/cvs/cv-1_v1.md",
		ChangeReason:  "Versión inicial",
	}
	if err := cvRepo.AddVersion(ctx, ver1); err != nil {
		t.Fatalf("error agregando versión: %v", err)
	}

	v2ID := uuid.NewString()
	ver2 := &domain.CVVersion{
		ID:            v2ID,
		CVID:          cvID,
		VersionNumber: 2,
		FilePath:      "data/cvs/cv-1_v2.md",
		ChangeReason:  "Optimizada por IA",
	}
	if err := cvRepo.AddVersion(ctx, ver2); err != nil {
		t.Fatalf("error agregando version 2: %v", err)
	}

	if err := cvRepo.SetFavoriteVersion(ctx, cvID, v2ID); err != nil {
		t.Fatalf("error marcando versión favorita: %v", err)
	}

	fetchedCV, err := cvRepo.GetByID(ctx, cvID)
	if err != nil {
		t.Fatalf("error obteniendo CV: %v", err)
	}
	if fetchedCV.FavoriteVersionID != v2ID {
		t.Errorf("esperaba favoriteVersionID %s, obtuve %s", v2ID, fetchedCV.FavoriteVersionID)
	}
	if len(fetchedCV.Versions) != 2 {
		t.Errorf("esperaba 2 versiones, obtuve %d", len(fetchedCV.Versions))
	}

	// 2. Test JobApplication & Requirements & Matching
	jobID := uuid.NewString()
	job := &domain.JobApplication{
		ID:               jobID,
		Company:          "Tech Corp",
		Role:             "Go Backend Engineer",
		JobPostingRaw:    "Buscamos desarrollador Go con 3 años de experiencia.",
		DetectedLanguage: domain.LangSpanish,
		TargetLanguage:   domain.LangSpanish,
		CVID:             cvID,
		CVVersionID:      v2ID,
		Status:           domain.StatusDraft,
		TotalATSScore:    85.5,
	}
	if err := jobRepo.Create(ctx, job); err != nil {
		t.Fatalf("error creando JobApplication: %v", err)
	}

	reqs := []domain.Requirement{
		{ID: uuid.NewString(), JobApplicationID: jobID, Name: "Go", Category: domain.CategoryHardSkill, Priority: domain.PriorityMandatory},
		{ID: uuid.NewString(), JobApplicationID: jobID, Name: "Docker", Category: domain.CategoryHardSkill, Priority: domain.PriorityDesirable},
	}
	if err := jobRepo.SaveRequirements(ctx, jobID, reqs); err != nil {
		t.Fatalf("error guardando requirements: %v", err)
	}

	savedReqs, err := jobRepo.GetRequirements(ctx, jobID)
	if err != nil || len(savedReqs) != 2 {
		t.Fatalf("error leyendo requirements: %v", err)
	}

	// 3. Test Prompts & Seeding
	defaults := []domain.Prompt{
		{ID: uuid.NewString(), Type: domain.PromptTypeJobAnalysis, Name: "Análisis de Oferta", Version: 1, Content: "Extrae requisitos."},
		{ID: uuid.NewString(), Type: domain.PromptTypeCVMatching, Name: "Matching CV", Version: 1, Content: "Compara requisitos."},
	}
	if err := promptRepo.SeedDefaults(ctx, defaults); err != nil {
		t.Fatalf("error seeding defaults: %v", err)
	}

	activePrompt, err := promptRepo.GetActive(ctx, domain.PromptTypeJobAnalysis)
	if err != nil {
		t.Fatalf("error obteniendo prompt activo: %v", err)
	}
	if activePrompt.Version != 1 {
		t.Errorf("esperaba versión 1, obtuve %d", activePrompt.Version)
	}

	// 4. Test Config
	cfg := &domain.ProviderConfig{
		ProviderName:    "gemini",
		EncryptedAPIKey: []byte("encrypted-secret"),
		DefaultModel:    "gemini-2.0-flash",
		IsActive:        true,
	}
	if err := configRepo.SaveConfig(ctx, cfg); err != nil {
		t.Fatalf("error guardando config: %v", err)
	}

	activeCfg, err := configRepo.GetActiveConfig(ctx)
	if err != nil {
		t.Fatalf("error obteniendo active config: %v", err)
	}
	if activeCfg.ProviderName != "gemini" {
		t.Errorf("esperaba provider gemini, obtuve %s", activeCfg.ProviderName)
	}

	// 5. Test Audit Log
	logEntry := &domain.AuditLog{
		Timestamp:    time.Now().UTC(),
		Action:       domain.ActionAnalyzeJob,
		Provider:     "gemini",
		Model:        "gemini-2.0-flash",
		ResultStatus: "SUCCESS",
		DetailsJSON:  `{"tokens": 120}`,
	}
	if err := auditRepo.Log(ctx, logEntry); err != nil {
		t.Fatalf("error guardando log: %v", err)
	}

	logs, err := auditRepo.List(ctx, 10)
	if err != nil || len(logs) == 0 {
		t.Fatalf("error listando logs de auditoría: %v", err)
	}
}
