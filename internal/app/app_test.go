package app

import (
	"context"
	"os"
	"testing"

	"github.com/jcarbajal/cv-checker/internal/domain"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/ai"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/diff"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/persistence/filestore"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/persistence/sqlite"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/security"
)

func TestFullApplicationFlow(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "app_flow_test_*")
	if err != nil {
		t.Fatalf("error creando tempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := tempDir + "/test.db"
	db, err := sqlite.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("error abriendo DB: %v", err)
	}
	defer db.Close()

	// Repositorios
	cvRepo := sqlite.NewSQLiteCVRepository(db)
	jobRepo := sqlite.NewSQLiteJobRepository(db)
	promptRepo := sqlite.NewSQLitePromptRepository(db)
	auditRepo := sqlite.NewSQLiteAuditRepository(db)
	configRepo := sqlite.NewSQLiteConfigRepository(db)

	fileStore, err := filestore.NewMarkdownStore(tempDir + "/cvs")
	if err != nil {
		t.Fatalf("error creando fileStore: %v", err)
	}

	encryptor, err := security.NewAESGCMEncryptor(tempDir + "/sec")
	if err != nil {
		t.Fatalf("error creando encryptor: %v", err)
	}

	diffEngine := diff.NewGitDiffEngine()
	aiFactory := ai.NewAIFactory(configRepo, encryptor)

	// Seed prompts
	ctx := context.Background()
	if err := promptRepo.SeedDefaults(ctx, ai.DefaultPrompts); err != nil {
		t.Fatalf("error sembrando prompts: %v", err)
	}

	// 1. Config Service & Provider
	cfgSvc := NewConfigSecurityService(configRepo, encryptor, auditRepo)
	if err := cfgSvc.SaveProviderConfig(ctx, "mock", "", "mock-model", "", true); err != nil {
		t.Fatalf("error configurando provider mock: %v", err)
	}

	// 2. CV Service: Crear CV inicial
	cvSvc := NewCVService(cvRepo, fileStore)
	initialMD := `# María López
Desarrollador Backend con experiencia en servicios web, bases de datos relacionales y contenedores Linux.
`
	cv, v1, err := cvSvc.CreateCV(ctx, "CV Backend Principal", "Mi CV base", domain.LangSpanish, initialMD)
	if err != nil {
		t.Fatalf("error creando CV: %v", err)
	}
	if !v1.IsFavorite {
		t.Errorf("la versión inicial v1 debe ser favorita")
	}

	// 3. Job Service: Crear Postulación
	jobSvc := NewJobService(jobRepo)
	posting := `Buscamos Ingeniero Backend Go con experiencia en bases de datos PostgreSQL y Docker.`
	job, err := jobSvc.CreateJob(ctx, "Fintech S.A.", "Go Developer", "https://fintech.example.com/job", posting, cv.ID, v1.ID, domain.LangSpanish)
	if err != nil {
		t.Fatalf("error creando job: %v", err)
	}

	// 4. Pipeline de Optimización
	pipe := NewOptimizationPipelineService(jobRepo, cvSvc, promptRepo, aiFactory, diffEngine, auditRepo)

	// Fase 1: Análisis de Oferta
	reqs, detectedLang, err := pipe.AnalyzeJobOffer(ctx, job.ID)
	if err != nil {
		t.Fatalf("error en AnalyzeJobOffer: %v", err)
	}
	if len(reqs) == 0 {
		t.Errorf("se esperaban requisitos extraídos")
	}
	if detectedLang != domain.LangSpanish {
		t.Errorf("idioma esperado español, obtenido %s", detectedLang)
	}

	// Fase 2 & 3: Matching y ATS Score
	atsScore, matches, recs, err := pipe.MatchCVAndScore(ctx, job.ID, v1.ID)
	if err != nil {
		t.Fatalf("error en MatchCVAndScore: %v", err)
	}
	if atsScore.TotalScore <= 0 {
		t.Errorf("total ATS score debe ser > 0, obtenido: %f", atsScore.TotalScore)
	}
	if len(matches) == 0 {
		t.Errorf("se esperaban matches evaluados")
	}
	if len(recs.Improvements) == 0 {
		t.Errorf("se esperaban recomendaciones generadas")
	}

	// Fase 4 & 5: Optimización con regla Anti-Invención, Diff y versionado
	optResult, err := pipe.OptimizeAndGenerateCV(ctx, job.ID, v1.ID, domain.LangSpanish)
	if err != nil {
		t.Fatalf("error en OptimizeAndGenerateCV: %v", err)
	}
	if optResult.NewVersion.VersionNumber != 2 {
		t.Errorf("esperaba versión 2 generada, obtuve %d", optResult.NewVersion.VersionNumber)
	}
	if optResult.DiffUnified == "" {
		t.Errorf("esperaba diff generado")
	}
	if len(optResult.Changes) == 0 {
		t.Errorf("esperaba cambios con trazabilidad")
	}

	// 5. Prompt Service: Versionado y Rollback
	promptSvc := NewPromptService(promptRepo, diffEngine, auditRepo)
	pActive, err := promptSvc.GetActivePrompt(ctx, domain.PromptTypeJobAnalysis)
	if err != nil {
		t.Fatalf("error obteniendo prompt activo: %v", err)
	}
	if pActive.Version != 1 {
		t.Errorf("versión activa esperada 1, obtenida %d", pActive.Version)
	}

	v2Prompt, err := promptSvc.CreateNewVersion(ctx, domain.PromptTypeJobAnalysis, "Prompt V2 Mejorado", "Contenido mejorado v2", true)
	if err != nil {
		t.Fatalf("error creando nueva versión de prompt: %v", err)
	}
	if v2Prompt.Version != 2 {
		t.Errorf("esperaba versión 2 de prompt, obtenida %d", v2Prompt.Version)
	}

	if err := promptSvc.RollbackToVersion(ctx, domain.PromptTypeJobAnalysis, 1); err != nil {
		t.Fatalf("error ejecutando rollback de prompt: %v", err)
	}
	pRestored, _ := promptSvc.GetActivePrompt(ctx, domain.PromptTypeJobAnalysis)
	if pRestored.Version != 1 {
		t.Errorf("esperaba prompt activo versión 1 tras rollback, obtuve %d", pRestored.Version)
	}

	// 6. Auditoría
	auditLogs, err := auditRepo.List(ctx, 50)
	if err != nil || len(auditLogs) < 4 {
		t.Fatalf("esperaba al menos 4 registros de auditoría, obtuve %d", len(auditLogs))
	}
}
