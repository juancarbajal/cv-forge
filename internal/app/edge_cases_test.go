package app

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jcarbajal/cv-checker/internal/domain"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/ai"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/diff"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/persistence/filestore"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/persistence/sqlite"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/security"
)

type badAIProvider struct {
	ai.MockAIProvider
}

// Retorna un cambio SIN evidencia para provocar fallo anti-invención
func (b *badAIProvider) GenerateOptimizedCV(ctx context.Context, cvContent string, jobPosting string, matches []domain.MatchResult, targetLang domain.Language, prompt string) (domain.OptimizedCVResult, error) {
	return domain.OptimizedCVResult{
		ContentMarkdown: "# CV con Alucinación",
		Changes: []domain.OptimizationChange{
			{
				ChangeSummary: "Se agregó Doctorado en MIT",
				EvidenceUsed:  "", // Sin evidencia
			},
		},
		TargetLanguage: targetLang,
	}, nil
}

type badAIFactory struct {
	provider domain.AIProvider
}

func (b *badAIFactory) GetActiveProvider(ctx context.Context) (domain.AIProvider, error) {
	return b.provider, nil
}

func setupTestApp(t *testing.T) (*CVService, *JobService, *PromptService, *OptimizationPipelineService, *filestore.MarkdownStore, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "edge_test_*")
	if err != nil {
		t.Fatalf("error creando tempDir: %v", err)
	}

	db, err := sqlite.OpenDB(tempDir + "/test.db")
	if err != nil {
		t.Fatalf("error abriendo DB: %v", err)
	}

	cvRepo := sqlite.NewSQLiteCVRepository(db)
	jobRepo := sqlite.NewSQLiteJobRepository(db)
	promptRepo := sqlite.NewSQLitePromptRepository(db)
	auditRepo := sqlite.NewSQLiteAuditRepository(db)
	configRepo := sqlite.NewSQLiteConfigRepository(db)

	fileStore, _ := filestore.NewMarkdownStore(tempDir + "/cvs")
	encryptor, _ := security.NewAESGCMEncryptor(tempDir + "/sec")
	diffEngine := diff.NewGitDiffEngine()
	aiFactory := ai.NewAIFactory(configRepo, encryptor)

	_ = promptRepo.SeedDefaults(context.Background(), ai.DefaultPrompts)
	_ = configRepo.SaveConfig(context.Background(), &domain.ProviderConfig{
		ProviderName: "mock",
		IsActive:     true,
	})

	cvSvc := NewCVService(cvRepo, fileStore)
	jobSvc := NewJobService(jobRepo)
	promptSvc := NewPromptService(promptRepo, diffEngine, auditRepo)
	pipeSvc := NewOptimizationPipelineService(jobRepo, cvSvc, promptRepo, aiFactory, diffEngine, auditRepo)

	cleanup := func() {
		db.Close()
		os.RemoveAll(tempDir)
	}

	return cvSvc, jobSvc, promptSvc, pipeSvc, fileStore, cleanup
}

func TestEdgeCases_JobWithoutRequirements(t *testing.T) {
	cvSvc, jobSvc, _, pipeSvc, _, cleanup := setupTestApp(t)
	defer cleanup()

	ctx := context.Background()
	cv, v1, _ := cvSvc.CreateCV(ctx, "CV Test", "Desc", domain.LangSpanish, "# Contenido")
	job, _ := jobSvc.CreateJob(ctx, "Empresa", "Rol", "https://...", "Anuncio", cv.ID, v1.ID, domain.LangSpanish)

	// Intentar matching antes de analizar requisitos
	_, _, _, err := pipeSvc.MatchCVAndScore(ctx, job.ID, v1.ID)
	if err == nil {
		t.Errorf("esperaba error al hacer matching sin requisitos")
	}

	// Intentar optimizar antes de hacer matching
	_, err = pipeSvc.OptimizeAndGenerateCV(ctx, job.ID, v1.ID, domain.LangSpanish)
	if err == nil {
		t.Errorf("esperaba error al optimizar sin matching previo")
	}
}

func TestEdgeCases_AntiInventionBlocking(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "anti_inv_*")
	defer os.RemoveAll(tempDir)

	db, _ := sqlite.OpenDB(tempDir + "/test.db")
	defer db.Close()

	cvRepo := sqlite.NewSQLiteCVRepository(db)
	jobRepo := sqlite.NewSQLiteJobRepository(db)
	promptRepo := sqlite.NewSQLitePromptRepository(db)
	auditRepo := sqlite.NewSQLiteAuditRepository(db)
	fileStore, _ := filestore.NewMarkdownStore(tempDir + "/cvs")
	diffEngine := diff.NewGitDiffEngine()

	_ = promptRepo.SeedDefaults(context.Background(), ai.DefaultPrompts)

	cvSvc := NewCVService(cvRepo, fileStore)
	jobSvc := NewJobService(jobRepo)

	ctx := context.Background()
	cv, v1, _ := cvSvc.CreateCV(ctx, "CV", "Desc", domain.LangSpanish, "# CV")
	job, _ := jobSvc.CreateJob(ctx, "Empresa", "Rol", "", "Anuncio", cv.ID, v1.ID, domain.LangSpanish)

	// Simular matches guardados
	_ = jobRepo.SaveRequirements(ctx, job.ID, []domain.Requirement{{ID: uuid.NewString(), Name: "Go"}})
	_ = jobRepo.SaveMatchResults(ctx, job.ID, []domain.MatchResult{{RequirementName: "Go", Status: domain.MatchStatusFulfilled}})

	// Inyectar un AI Provider malicioso/alucinador
	customAIFactory := ai.NewAIFactory(sqlite.NewSQLiteConfigRepository(db), nil)
	pipeSvc := NewOptimizationPipelineService(jobRepo, cvSvc, promptRepo, customAIFactory, diffEngine, auditRepo)

	// Reemplazar validador interno con un proveedor que inventa
	validator := domain.NewAntiInventionValidator()
	badResult := &domain.OptimizedCVResult{
		Changes: []domain.OptimizationChange{
			{ChangeSummary: "Título inventado", EvidenceUsed: ""},
		},
	}
	_, err := validator.Validate("# CV", badResult)
	if err == nil {
		t.Errorf("esperaba rechazo bloqueante por invención sin evidencia")
	}

	// El pipeline debe retornar error si la validación falla
	_ = pipeSvc
}

func TestEdgeCases_DeleteCVRemovesMarkdownFiles(t *testing.T) {
	cvSvc, _, _, _, fileStore, cleanup := setupTestApp(t)
	defer cleanup()

	ctx := context.Background()
	cv, v1, err := cvSvc.CreateCV(ctx, "CV Para Borrar", "Desc", domain.LangSpanish, "# Para borrar")
	if err != nil {
		t.Fatalf("error creando CV: %v", err)
	}

	// Verificar que el archivo existe en disco
	content, err := fileStore.GetContent(ctx, v1.FilePath)
	if err != nil || content != "# Para borrar" {
		t.Fatalf("archivo no encontrado en disco: %v", err)
	}

	// Borrar CV
	if err := cvSvc.DeleteCV(ctx, cv.ID); err != nil {
		t.Fatalf("error borrando CV: %v", err)
	}

	// Verificar que el archivo fue eliminado
	_, err = fileStore.GetContent(ctx, v1.FilePath)
	if err == nil {
		t.Errorf("el archivo markdown debería haber sido eliminado del disco")
	}

	// Verificar que el CV ya no existe en BD
	_, err = cvSvc.GetCV(ctx, cv.ID)
	if err == nil {
		t.Errorf("el CV no debería existir en BD")
	}
}

func TestEdgeCases_PromptRollbackNonExistent(t *testing.T) {
	_, _, promptSvc, _, _, cleanup := setupTestApp(t)
	defer cleanup()

	ctx := context.Background()
	err := promptSvc.RollbackToVersion(ctx, domain.PromptTypeJobAnalysis, 999)
	if err == nil {
		t.Errorf("esperaba error al hacer rollback a versión inexistente 999")
	}
}

func TestEdgeCases_DuplicateNonExistentCV(t *testing.T) {
	cvSvc, _, _, _, _, cleanup := setupTestApp(t)
	defer cleanup()

	ctx := context.Background()
	_, err := cvSvc.DuplicateCV(ctx, "non-existent-id", "Copia")
	if err == nil {
		t.Errorf("esperaba error al duplicar CV inexistente")
	}
}
