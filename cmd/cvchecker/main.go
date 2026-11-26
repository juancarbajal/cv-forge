package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jcarbajal/cv-checker/internal/app"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/ai"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/diff"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/persistence/filestore"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/persistence/sqlite"
	"github.com/jcarbajal/cv-checker/internal/infrastructure/security"
	"github.com/jcarbajal/cv-checker/internal/ui"
)

func main() {
	// 1. Directorios base de datos y almacenamiento local
	baseDataDir := "data"
	cvsDir := filepath.Join(baseDataDir, "cvs")
	dbPath := filepath.Join(baseDataDir, "cvchecker.db")
	secDir := filepath.Join(baseDataDir, "security")

	for _, dir := range []string{cvsDir, secDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Error inicializando directorios locales: %v", err)
		}
	}

	// 2. Base de datos SQLite y migraciones automáticas
	db, err := sqlite.OpenDB(dbPath)
	if err != nil {
		log.Fatalf("Error conectando a SQLite: %v", err)
	}
	defer db.Close()

	// 3. Repositorios de persistencia
	cvRepo := sqlite.NewSQLiteCVRepository(db)
	jobRepo := sqlite.NewSQLiteJobRepository(db)
	promptRepo := sqlite.NewSQLitePromptRepository(db)
	auditRepo := sqlite.NewSQLiteAuditRepository(db)
	configRepo := sqlite.NewSQLiteConfigRepository(db)

	fileStore, err := filestore.NewMarkdownStore(cvsDir)
	if err != nil {
		log.Fatalf("Error inicializando almacenamiento de CVs: %v", err)
	}

	// 4. Seguridad y Cifrado AES-256 GCM
	encryptor, err := security.NewAESGCMEncryptor(secDir)
	if err != nil {
		log.Fatalf("Error inicializando motor de cifrado: %v", err)
	}

	// 5. Sembrado de Prompts por defecto (Seed Defaults)
	ctx := context.Background()
	if err := promptRepo.SeedDefaults(ctx, ai.DefaultPrompts); err != nil {
		fmt.Printf("Aviso: no se pudieron sembrar los prompts iniciales: %v\n", err)
	}

	// 6. Servicios de Aplicación
	diffEngine := diff.NewGitDiffEngine()
	aiFactory := ai.NewAIFactory(configRepo, encryptor)

	cvService := app.NewCVService(cvRepo, fileStore)
	jobService := app.NewJobService(jobRepo)
	promptService := app.NewPromptService(promptRepo, diffEngine, auditRepo)
	configService := app.NewConfigSecurityService(configRepo, encryptor, auditRepo)
	auditService := app.NewAuditService(auditRepo)
	pipeService := app.NewOptimizationPipelineService(jobRepo, cvService, promptRepo, aiFactory, diffEngine, auditRepo)

	// 7. Lanzamiento de la Interfaz Gráfica Fyne
	application := ui.NewApp(cvService, jobService, pipeService, promptService, configService, auditService)
	application.Run()
}
