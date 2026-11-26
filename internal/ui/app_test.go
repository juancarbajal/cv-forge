package ui

import (
	"context"
	"os"
	"testing"

	"fyne.io/fyne/v2/test"

	appservice "github.com/jcarbajal/cv-checker/internal/app"
	"github.com/jcarbajal/cv-checker/internal/domain"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/ai"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/diff"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/persistence/filestore"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/persistence/sqlite"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/security"
	"github.com/jcarbajal/cv-checker/internal/ui/views"
)

func TestUIViewsInitialization(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ui_test_*")
	if err != nil {
		t.Fatalf("error creando tempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	db, err := sqlite.OpenDB(tempDir + "/test.db")
	if err != nil {
		t.Fatalf("error abriendo DB: %v", err)
	}
	defer db.Close()

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

	cvSvc := appservice.NewCVService(cvRepo, fileStore)
	jobSvc := appservice.NewJobService(jobRepo)
	promptSvc := appservice.NewPromptService(promptRepo, diffEngine, auditRepo)
	configSvc := appservice.NewConfigSecurityService(configRepo, encryptor, auditRepo)
	auditSvc := appservice.NewAuditService(auditRepo)
	pipeSvc := appservice.NewOptimizationPipelineService(jobRepo, cvSvc, promptRepo, aiFactory, diffEngine, auditRepo)

	// Crear CV inicial de prueba
	_, _, err = cvSvc.CreateCV(context.Background(), "CV Test", "Desc", domain.LangSpanish, "# Test CV")
	if err != nil {
		t.Fatalf("error creando CV: %v", err)
	}

	// Inicializar test app de Fyne (headless / en memoria)
	testApp := test.NewApp()
	testWin := testApp.NewWindow("Test")

	// 1. Validar Vista de CVs
	cvView := views.NewCVView(cvSvc, testWin)
	if cvView == nil {
		t.Fatalf("error inicializando CVView")
	}

	// 2. Validar Vista de Postulaciones
	jobView := views.NewJobView(jobSvc, cvSvc, pipeSvc, testWin)
	if jobView == nil {
		t.Fatalf("error inicializando JobView")
	}

	// 3. Validar Vista de Prompts
	promptView := views.NewPromptView(promptSvc, testWin)
	if promptView == nil {
		t.Fatalf("error inicializando PromptView")
	}

	// 4. Validar Vista de Configuración
	settingsView := views.NewSettingsView(configSvc, testWin)
	if settingsView == nil {
		t.Fatalf("error inicializando SettingsView")
	}

	// 5. Validar Vista de Auditoría
	auditView := views.NewAuditView(auditSvc, testWin)
	if auditView == nil {
		t.Fatalf("error inicializando AuditView")
	}
}
