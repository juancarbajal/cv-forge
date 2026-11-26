package views

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/jcarbajal/cv-checker/internal/app"
	"github.com/jcarbajal/cv-checker/internal/domain"
)

type JobView struct {
	jobService *app.JobService
	cvService  *app.CVService
	pipeline   *app.OptimizationPipelineService
	win        fyne.Window

	jobList     *widget.List
	jobs        []domain.JobApplication
	selectedJob *domain.JobApplication

	// Formulario Oferta
	companyEntry *widget.Entry
	roleEntry    *widget.Entry
	urlEntry     *widget.Entry
	postingEntry *widget.Entry
	cvSelect     *widget.Select
	cvMap        map[string]domain.CV
	langSelect   *widget.Select

	// Vista Análisis & Matching
	reqsContainer    *fyne.Container
	atsScoreLabel    *widget.Label
	atsDetailLabel   *widget.Label
	matchesContainer *fyne.Container
	recsContainer    *fyne.Container

	// Vista Optimización & Diff
	diffViewer       *widget.Entry
	changesContainer *fyne.Container
	agentTraceLabel  *widget.Label
	gapAnswers       []domain.GapAnswer
}

func NewJobView(jobService *app.JobService, cvService *app.CVService, pipeline *app.OptimizationPipelineService, win fyne.Window) fyne.CanvasObject {
	v := &JobView{
		jobService: jobService,
		cvService:  cvService,
		pipeline:   pipeline,
		win:        win,
		cvMap:      make(map[string]domain.CV),
	}
	return v.buildUI()
}

func (v *JobView) buildUI() fyne.CanvasObject {
	// 1. Panel Izquierdo: Lista de Postulaciones
	v.jobList = widget.NewList(
		func() int { return len(v.jobs) },
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabelWithStyle("Empresa - Cargo", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel("Estado | ATS: 0%"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(v.jobs) {
				j := v.jobs[id]
				box := obj.(*fyne.Container)
				box.Objects[0].(*widget.Label).SetText(fmt.Sprintf("%s - %s", j.Company, j.Role))
				box.Objects[1].(*widget.Label).SetText(fmt.Sprintf("Estado: %s | ATS: %.1f%%", j.Status, j.TotalATSScore))
			}
		},
	)
	v.jobList.OnSelected = func(id widget.ListItemID) {
		if id < len(v.jobs) {
			v.selectJob(v.jobs[id].ID)
		}
	}

	btnNewJob := widget.NewButton("Nueva Postulación", func() {
		v.showCreateJobDialog()
	})

	leftPanel := container.NewBorder(
		widget.NewLabelWithStyle("Postulaciones Laborales", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		btnNewJob,
		nil,
		nil,
		v.jobList,
	)

	// 2. Panel Derecho: SubTabs (Oferta, Análisis ATS, Optimización Diff)
	// --- Tab 1: Datos de la Oferta ---
	v.companyEntry = widget.NewEntry()
	v.companyEntry.SetPlaceHolder("Ej: MercadoLibre, Globant, Google...")

	v.roleEntry = widget.NewEntry()
	v.roleEntry.SetPlaceHolder("Ej: Senior Backend Engineer Go...")

	v.urlEntry = widget.NewEntry()
	v.urlEntry.SetPlaceHolder("https://empresa.com/careers/job-123 (opcional)")

	v.cvSelect = widget.NewSelect([]string{}, nil)
	v.langSelect = widget.NewSelect([]string{"Español (es)", "Inglés (en)", "Portugués (pt)"}, nil)

	v.postingEntry = widget.NewMultiLineEntry()
	v.postingEntry.SetPlaceHolder("Pegue aquí el texto completo del anuncio o descripción de la vacante laboral...")

	btnSaveJob := widget.NewButton("Actualizar Datos de Postulación", func() {
		if v.selectedJob == nil {
			return
		}
		v.selectedJob.Company = v.companyEntry.Text
		v.selectedJob.Role = v.roleEntry.Text
		v.selectedJob.URL = v.urlEntry.Text
		v.selectedJob.JobPostingRaw = v.postingEntry.Text

		if selectedCV, ok := v.cvMap[v.cvSelect.Selected]; ok {
			v.selectedJob.CVID = selectedCV.ID
			v.selectedJob.CVVersionID = selectedCV.FavoriteVersionID
		}

		err := v.jobService.UpdateJob(context.Background(), v.selectedJob)
		if err != nil {
			dialog.ShowError(err, v.win)
			return
		}
		v.reloadJobs()
		dialog.ShowInformation("Éxito", "Postulación actualizada correctamente", v.win)
	})

	formOferta := widget.NewForm(
		widget.NewFormItem("Empresa:", v.companyEntry),
		widget.NewFormItem("Cargo / Posición:", v.roleEntry),
		widget.NewFormItem("URL de la Oferta:", v.urlEntry),
		widget.NewFormItem("CV Seleccionado:", v.cvSelect),
		widget.NewFormItem("Idioma Objetivo:", v.langSelect),
	)

	postingContainer := container.NewBorder(
		widget.NewLabelWithStyle("Texto del Anuncio Laboral / Descripción de la Vacante:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		v.postingEntry,
	)

	tabOferta := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Datos Principales de la Oferta", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			formOferta,
			btnSaveJob,
		),
		nil, nil, nil,
		postingContainer,
	)

	// --- Tab 2: Análisis & Score ATS ---
	v.atsScoreLabel = widget.NewLabelWithStyle("ATS Score: No evaluado", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	v.atsDetailLabel = widget.NewLabel("Hard Skills: - | Experiencia: - | Keywords: - | Educación: - | Idioma: -")
	v.reqsContainer = container.NewVBox()
	v.matchesContainer = container.NewVBox()
	v.recsContainer = container.NewVBox()

	btnAnalyze := widget.NewButton("1. Analizar Oferta (Extraer Requisitos)", func() {
		if v.selectedJob == nil {
			return
		}
		RunAsync(v.win, "Analizando Oferta...", "Extrayendo requisitos y detectando idioma con IA...", 50*time.Second, func(ctx context.Context) error {
			_, _, err := v.pipeline.AnalyzeJobOffer(ctx, v.selectedJob.ID)
			return err
		}, func(err error) {
			if err != nil {
				dialog.ShowError(err, v.win)
				return
			}
			v.selectJob(v.selectedJob.ID)
			dialog.ShowInformation("Completado", "Requisitos de la oferta analizados correctamente", v.win)
		})
	})

	btnMatchScore := widget.NewButton("2. Evaluar Matching & Calcular ATS", func() {
		if v.selectedJob == nil {
			return
		}
		RunAsync(v.win, "Evaluando CV...", "Comparando requisitos y calculando puntuación ATS...", 50*time.Second, func(ctx context.Context) error {
			_, _, _, err := v.pipeline.MatchCVAndScore(ctx, v.selectedJob.ID, v.selectedJob.CVVersionID)
			return err
		}, func(err error) {
			if err != nil {
				dialog.ShowError(err, v.win)
				return
			}
			v.selectJob(v.selectedJob.ID)
			dialog.ShowInformation("Completado", "Matching y Score ATS calculados", v.win)
		})
	})

	analysisActionButtons := container.NewHBox(btnAnalyze, btnMatchScore)
	atsCard := widget.NewCard("Resultado ATS Explicable", "", container.NewVBox(
		analysisActionButtons,
		v.atsScoreLabel,
		v.atsDetailLabel,
	))

	tabAnalisis := container.NewBorder(
		atsCard,
		nil,
		nil,
		nil,
		container.NewVScroll(container.NewVBox(
			widget.NewLabelWithStyle("Requisitos Detectados en la Oferta:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			v.reqsContainer,
			widget.NewLabelWithStyle("Matching y Evidencia Encontrada:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			v.matchesContainer,
			widget.NewLabelWithStyle("Recomendaciones y Keywords:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			v.recsContainer,
		)),
	)

	// --- Tab 3: Optimización & Diff ---
	v.agentTraceLabel = widget.NewLabel("Estado Agentes: Sin ejecutar")
	v.agentTraceLabel.Wrapping = fyne.TextWrapWord

	btnGapInterview := widget.NewButton("2b. Responder Preguntas de Gaps (Opcional)", func() {
		if v.selectedJob == nil {
			return
		}
		var questions []domain.GapQuestion
		RunAsync(v.win, "Analizando Gaps...", "El Agente Entrevistador está preparando preguntas sobre requisitos faltantes...", 55*time.Second, func(ctx context.Context) error {
			qs, err := v.pipeline.GenerateGapQuestions(ctx, v.selectedJob.ID, v.selectedJob.CVVersionID)
			if err != nil {
				return err
			}
			questions = qs
			return nil
		}, func(err error) {
			if err != nil {
				dialog.ShowError(err, v.win)
				return
			}
			if len(questions) == 0 {
				dialog.ShowInformation("Sin Gaps", "No se detectaron requisitos faltantes o parciales para entrevistar.", v.win)
				return
			}
			entries := make([]*widget.Entry, len(questions))
			formItems := make([]*widget.FormItem, len(questions))
			for i, q := range questions {
				e := widget.NewMultiLineEntry()
				e.SetPlaceHolder(q.Question)
				entries[i] = e
				formItems[i] = widget.NewFormItem(fmt.Sprintf("%s:\n%s", q.RequirementName, q.Question), e)
			}
			dialog.ShowForm("Entrevista de Gaps (Agente Entrevistador)", "Guardar Respuestas", "Cancelar", formItems, func(ok bool) {
				if !ok {
					return
				}
				var answers []domain.GapAnswer
				for i, q := range questions {
					if entries[i].Text != "" {
						answers = append(answers, domain.GapAnswer{
							RequirementName: q.RequirementName,
							Question:        q.Question,
							Answer:          entries[i].Text,
						})
					}
				}
				v.gapAnswers = answers
				v.agentTraceLabel.SetText(fmt.Sprintf("Estado Agentes: %d respuestas de gaps listas para la optimización.", len(answers)))
			}, v.win)
		})
	})

	btnOptimize := widget.NewButton("3. Generar CV Optimizado con IA (Agentes + Anti-Invención)", func() {
		if v.selectedJob == nil {
			return
		}
		RunAsync(v.win, "Optimizando CV...", "Ejecutando ReflectionOptimizer + CriticAgent (máx. 2 iteraciones)...", 150*time.Second, func(ctx context.Context) error {
			res, err := v.pipeline.OptimizeWithContext(ctx, v.selectedJob.ID, v.selectedJob.CVVersionID, v.selectedJob.TargetLanguage, v.gapAnswers)
			if err != nil {
				return err
			}
			fyne.Do(func() {
				v.diffViewer.SetText(res.DiffUnified)
				v.renderChanges(res.Changes)
				v.agentTraceLabel.SetText(fmt.Sprintf("Auditoría Agéntica (%d pasos): %s", len(res.AgentTrace), res.CriticSummary))
			})
			return nil
		}, func(err error) {
			if err != nil {
				dialog.ShowError(err, v.win)
				return
			}
			v.selectJob(v.selectedJob.ID)
			dialog.ShowInformation("Éxito", "Nueva versión del CV generada, auditada y guardada en el historial", v.win)
		})
	})

	v.diffViewer = widget.NewMultiLineEntry()
	v.diffViewer.SetPlaceHolder("El diff estilo Git aparecerá aquí tras ejecutar la optimización...")
	v.changesContainer = container.NewVBox()

	diffSection := container.NewBorder(
		widget.NewLabelWithStyle("Diferencias Estilo Git (Original vs Optimizado)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		v.diffViewer,
	)
	changesSection := container.NewBorder(
		widget.NewLabelWithStyle("Trazabilidad de Cambios Realizados", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		container.NewVScroll(v.changesContainer),
	)

	tabOptSplit := container.NewHSplit(diffSection, changesSection)
	tabOptSplit.SetOffset(0.6)

	tabOptimizacion := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Genera un CV optimizado adaptado específicamente a los requisitos de la vacante sin inventar información."),
			container.NewHBox(btnGapInterview, btnOptimize),
			v.agentTraceLabel,
		),
		nil, nil, nil,
		tabOptSplit,
	)

	appTabs := container.NewAppTabs(
		container.NewTabItem("Oferta Laboral", tabOferta),
		container.NewTabItem("Análisis & Score ATS", tabAnalisis),
		container.NewTabItem("Optimización & Diff", tabOptimizacion),
	)

	mainSplit := container.NewHSplit(leftPanel, appTabs)
	mainSplit.SetOffset(0.3)

	v.loadCVOptions()
	v.reloadJobs()
	return mainSplit
}

func (v *JobView) loadCVOptions() {
	cvs, err := v.cvService.ListCVs(context.Background(), "")
	if err == nil {
		var titles []string
		v.cvMap = make(map[string]domain.CV)
		for _, c := range cvs {
			v.cvMap[c.Title] = c
			titles = append(titles, c.Title)
		}
		v.cvSelect.Options = titles
		v.cvSelect.Refresh()
	}
}

func (v *JobView) reloadJobs() {
	jobs, err := v.jobService.ListJobs(context.Background())
	if err == nil {
		v.jobs = jobs
		v.jobList.Refresh()
	}
}

func (v *JobView) selectJob(id string) {
	job, reqs, matches, score, changes, err := v.jobService.GetJobHistory(context.Background(), id)
	if err != nil {
		return
	}
	v.selectedJob = job

	v.companyEntry.SetText(job.Company)
	v.roleEntry.SetText(job.Role)
	v.urlEntry.SetText(job.URL)
	v.postingEntry.SetText(job.JobPostingRaw)

	// Seleccionar CV en combobox
	for title, c := range v.cvMap {
		if c.ID == job.CVID {
			v.cvSelect.SetSelected(title)
			break
		}
	}

	// Renderizar ATS Score
	if score != nil {
		v.atsScoreLabel.SetText(fmt.Sprintf("ATS Score Total: %.1f / 100", score.TotalScore))
		v.atsDetailLabel.SetText(fmt.Sprintf(
			"Hard Skills (35%%): %.1f | Experiencia (25%%): %.1f | Keywords (20%%): %.1f | Educación (10%%): %.1f | Idioma (10%%): %.1f",
			score.HardSkillsScore, score.ExperienceScore, score.KeywordsScore, score.EducationScore, score.LanguageScore,
		))
	} else {
		v.atsScoreLabel.SetText("ATS Score: No evaluado")
		v.atsDetailLabel.SetText("Hard Skills: - | Experiencia: - | Keywords: - | Educación: - | Idioma: -")
	}

	// Renderizar Requisitos
	v.reqsContainer.Objects = nil
	for _, r := range reqs {
		v.reqsContainer.Add(widget.NewLabel(fmt.Sprintf("• [%s | %s] %s", r.Category, r.Priority, r.Name)))
	}
	v.reqsContainer.Refresh()

	// Renderizar Matches
	v.matchesContainer.Objects = nil
	for _, m := range matches {
		v.matchesContainer.Add(widget.NewLabel(fmt.Sprintf("• %s -> %s (Confianza: %.0f%%)\n  Evidencia: %s\n  Explicación: %s",
			m.RequirementName, m.Status, m.Confidence*100, m.Evidence, m.Explanation)))
	}
	v.matchesContainer.Refresh()

	// Renderizar Cambios
	v.renderChanges(changes)
}

func (v *JobView) renderChanges(changes []domain.OptimizationChange) {
	v.changesContainer.Objects = nil
	for _, c := range changes {
		card := widget.NewCard(
			c.ChangeSummary,
			fmt.Sprintf("Impacto ATS: +%.1f pts | Requisito: %s", c.ExpectedATSImpact, c.RelatedRequirementName),
			container.NewVBox(
				widget.NewLabel(fmt.Sprintf("Motivo: %s", c.Reason)),
				widget.NewLabel(fmt.Sprintf("Evidencia Origen: %s", c.EvidenceUsed)),
			),
		)
		v.changesContainer.Add(card)
	}
	v.changesContainer.Refresh()
}

func (v *JobView) showCreateJobDialog() {
	v.loadCVOptions()

	companyEntry := widget.NewEntry()
	companyEntry.SetPlaceHolder("Ej: MercadoLibre, Globant, Google...")

	roleEntry := widget.NewEntry()
	roleEntry.SetPlaceHolder("Ej: Senior Backend Engineer Go...")

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://empresa.com/careers/job-123 (opcional)")

	cvSel := widget.NewSelect(v.cvSelect.Options, nil)
	if len(v.cvSelect.Options) > 0 {
		cvSel.SetSelectedIndex(0)
	}

	langSel := widget.NewSelect([]string{"Español (es)", "Inglés (en)", "Portugués (pt)"}, nil)
	langSel.SetSelected("Español (es)")

	postingEntry := widget.NewMultiLineEntry()
	postingEntry.SetPlaceHolder("Pegue aquí el texto completo o requisitos de la vacante...")

	formItems := []*widget.FormItem{
		widget.NewFormItem("Empresa Contratante:", companyEntry),
		widget.NewFormItem("Cargo / Posición:", roleEntry),
		widget.NewFormItem("URL del Puesto (Opcional):", urlEntry),
		widget.NewFormItem("CV Base a Optimizar:", cvSel),
		widget.NewFormItem("Idioma Objetivo de Salida:", langSel),
		widget.NewFormItem("Texto del Anuncio Laboral:", postingEntry),
	}

	dialog.ShowForm("Crear Nueva Postulación", "Crear", "Cancelar", formItems, func(ok bool) {
		if !ok || companyEntry.Text == "" || roleEntry.Text == "" || postingEntry.Text == "" {
			return
		}

		selectedCV, exists := v.cvMap[cvSel.Selected]
		if !exists {
			dialog.ShowError(fmt.Errorf("debe seleccionar un CV válido"), v.win)
			return
		}

		var targetLang domain.Language = domain.LangSpanish
		if langSel.Selected == "Inglés (en)" {
			targetLang = domain.LangEnglish
		} else if langSel.Selected == "Portugués (pt)" {
			targetLang = domain.LangPortuguese
		}

		job, err := v.jobService.CreateJob(context.Background(),
			companyEntry.Text, roleEntry.Text, urlEntry.Text, postingEntry.Text,
			selectedCV.ID, selectedCV.FavoriteVersionID, targetLang,
		)
		if err != nil {
			dialog.ShowError(err, v.win)
			return
		}
		v.reloadJobs()
		v.selectJob(job.ID)
	}, v.win)
}
