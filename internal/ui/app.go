package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	appservice "github.com/jcarbajal/cv-checker/internal/app"
	"github.com/jcarbajal/cv-checker/internal/ui/views"
)

type App struct {
	fyneApp   fyne.App
	win       fyne.Window
	cvSvc     *appservice.CVService
	jobSvc    *appservice.JobService
	pipeSvc   *appservice.OptimizationPipelineService
	promptSvc *appservice.PromptService
	configSvc *appservice.ConfigSecurityService
	auditSvc  *appservice.AuditService
}

func NewApp(
	cvSvc *appservice.CVService,
	jobSvc *appservice.JobService,
	pipeSvc *appservice.OptimizationPipelineService,
	promptSvc *appservice.PromptService,
	configSvc *appservice.ConfigSecurityService,
	auditSvc *appservice.AuditService,
) *App {
	a := app.NewWithID("com.cvchecker.app")
	win := a.NewWindow("CV Checker - Optimizador Inteligente de CVs ATS")
	win.Resize(fyne.NewSize(1200, 780))

	return &App{
		fyneApp:   a,
		win:       win,
		cvSvc:     cvSvc,
		jobSvc:    jobSvc,
		pipeSvc:   pipeSvc,
		promptSvc: promptSvc,
		configSvc: configSvc,
		auditSvc:  auditSvc,
	}
}

func (a *App) Run() {
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Gestión de CVs", theme.DocumentIcon(), views.NewCVView(a.cvSvc, a.win)),
		container.NewTabItemWithIcon("Postulaciones & ATS", theme.NavigateNextIcon(), views.NewJobView(a.jobSvc, a.cvSvc, a.pipeSvc, a.win)),
		container.NewTabItemWithIcon("Gestión de Prompts", theme.MenuIcon(), views.NewPromptView(a.promptSvc, a.win)),
		container.NewTabItemWithIcon("Configuración IA", theme.SettingsIcon(), views.NewSettingsView(a.configSvc, a.win)),
		container.NewTabItemWithIcon("Auditoría", theme.InfoIcon(), views.NewAuditView(a.auditSvc, a.win)),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	a.win.SetContent(tabs)
	a.win.ShowAndRun()
}
