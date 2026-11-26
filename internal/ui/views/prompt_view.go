package views

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/jcarbajal/cv-checker/internal/app"
	"github.com/jcarbajal/cv-checker/internal/domain"
)

type PromptView struct {
	promptService *app.PromptService
	win           fyne.Window

	typeSelect     *widget.Select
	versionList    *widget.List
	versions       []domain.Prompt
	selectedPrompt *domain.Prompt

	nameEntry   *widget.Entry
	editorEntry *widget.Entry
	makeActive  *widget.Check
	diffEntry   *widget.Entry
}

func NewPromptView(promptService *app.PromptService, win fyne.Window) fyne.CanvasObject {
	v := &PromptView{
		promptService: promptService,
		win:           win,
	}
	return v.buildUI()
}

func (v *PromptView) buildUI() fyne.CanvasObject {
	promptTypes := []string{
		string(domain.PromptTypeJobAnalysis),
		string(domain.PromptTypeCVMatching),
		string(domain.PromptTypeATSScoring),
		string(domain.PromptTypeCVOptimization),
		string(domain.PromptTypeTranslation),
	}

	v.typeSelect = widget.NewSelect(promptTypes, func(selected string) {
		v.reloadVersions(domain.PromptType(selected))
	})

	v.versionList = widget.NewList(
		func() int { return len(v.versions) },
		func() fyne.CanvasObject {
			return widget.NewLabel("v1 - Nombre ★")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(v.versions) {
				p := v.versions[id]
				active := ""
				if p.IsActive {
					active = " ★ ACTIVA"
				}
				obj.(*widget.Label).SetText(fmt.Sprintf("v%d (%s)%s", p.Version, p.Name, active))
			}
		},
	)
	v.versionList.OnSelected = func(id widget.ListItemID) {
		if id < len(v.versions) {
			v.selectPrompt(&v.versions[id])
		}
	}

	btnRollback := widget.NewButton("Restaurar Versión Seleccionada (Rollback)", func() {
		if v.selectedPrompt == nil {
			dialog.ShowInformation("Atención", "Seleccione una versión para restaurar", v.win)
			return
		}
		err := v.promptService.RollbackToVersion(context.Background(), v.selectedPrompt.Type, v.selectedPrompt.Version)
		if err != nil {
			dialog.ShowError(err, v.win)
			return
		}
		v.reloadVersions(v.selectedPrompt.Type)
		dialog.ShowInformation("Éxito", fmt.Sprintf("Versión %d activada correctamente", v.selectedPrompt.Version), v.win)
	})

	btnCompare := widget.NewButton("Comparar con Versión Activa (Diff)", func() {
		if v.selectedPrompt == nil {
			return
		}
		activeP, err := v.promptService.GetActivePrompt(context.Background(), v.selectedPrompt.Type)
		if err != nil {
			dialog.ShowError(err, v.win)
			return
		}
		diffStr, err := v.promptService.CompareVersions(context.Background(), v.selectedPrompt.Type, v.selectedPrompt.Version, activeP.Version)
		if err != nil {
			dialog.ShowError(err, v.win)
			return
		}
		v.diffEntry.SetText(diffStr)
	})

	leftPanel := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Catálogo de Prompts", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Seleccionar Tipo de Prompt:"),
			v.typeSelect,
			widget.NewLabelWithStyle("Versiones Registradas:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		),
		container.NewVBox(btnRollback, btnCompare),
		nil, nil,
		v.versionList,
	)

	// Editor y Diff
	v.nameEntry = widget.NewEntry()
	v.nameEntry.SetPlaceHolder("Ej: Optimización de viñetas con métricas de impacto...")

	v.editorEntry = widget.NewMultiLineEntry()
	v.editorEntry.SetPlaceHolder("Instrucciones del prompt del sistema para el modelo de IA...")

	v.makeActive = widget.NewCheck("Establecer inmediatamente como versión activa", nil)
	v.makeActive.SetChecked(true)

	btnSave := widget.NewButton("Guardar como Nueva Versión (v+1)", func() {
		if v.typeSelect.Selected == "" {
			return
		}
		name := v.nameEntry.Text
		if name == "" {
			name = "Ajuste de prompt"
		}
		pType := domain.PromptType(v.typeSelect.Selected)
		newP, err := v.promptService.CreateNewVersion(context.Background(), pType, name, v.editorEntry.Text, v.makeActive.Checked)
		if err != nil {
			dialog.ShowError(err, v.win)
			return
		}
		v.reloadVersions(pType)
		v.selectPrompt(newP)
		dialog.ShowInformation("Éxito", fmt.Sprintf("Nueva versión v%d creada", newP.Version), v.win)
	})

	v.diffEntry = widget.NewMultiLineEntry()
	v.diffEntry.SetPlaceHolder("El diff de cambios entre versiones se mostrará aquí...")

	promptForm := widget.NewForm(
		widget.NewFormItem("Nombre / Resumen de la Versión:", v.nameEntry),
	)

	editorBox := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Editor de Prompt", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			promptForm,
			v.makeActive,
			widget.NewLabelWithStyle("Instrucciones del Prompt del Sistema (System Prompt):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		),
		btnSave,
		nil, nil,
		v.editorEntry,
	)

	diffBox := container.NewBorder(
		widget.NewLabelWithStyle("Visor Diff de Prompts", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		v.diffEntry,
	)

	rightSplit := container.NewHSplit(editorBox, diffBox)
	rightSplit.SetOffset(0.6)

	mainSplit := container.NewHSplit(leftPanel, rightSplit)
	mainSplit.SetOffset(0.3)

	v.typeSelect.SetSelected(promptTypes[0])
	return mainSplit
}

func (v *PromptView) reloadVersions(pType domain.PromptType) {
	vers, err := v.promptService.ListVersions(context.Background(), pType)
	if err == nil {
		v.versions = vers
		v.versionList.Refresh()
		for i := range vers {
			if vers[i].IsActive {
				v.selectPrompt(&vers[i])
				break
			}
		}
	}
}

func (v *PromptView) selectPrompt(p *domain.Prompt) {
	v.selectedPrompt = p
	v.nameEntry.SetText(p.Name)
	v.editorEntry.SetText(p.Content)
}
