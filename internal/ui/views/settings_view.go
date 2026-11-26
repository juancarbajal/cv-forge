package views

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/jcarbajal/cv-checker/internal/app"
	"github.com/jcarbajal/cv-checker/internal/domain"
)

type SettingsView struct {
	configService *app.ConfigSecurityService
	win           fyne.Window

	providerSelect *widget.Select
	apiKeyEntry    *widget.Entry
	modelEntry     *widget.Entry
	baseURLEntry   *widget.Entry
	activeCheck    *widget.Check
	statusLabel    *widget.Label

	configList *widget.List
	configs    []domain.ProviderConfig
}

func NewSettingsView(configService *app.ConfigSecurityService, win fyne.Window) fyne.CanvasObject {
	v := &SettingsView{
		configService: configService,
		win:           win,
	}
	return v.buildUI()
}

func (v *SettingsView) buildUI() fyne.CanvasObject {
	v.providerSelect = widget.NewSelect([]string{"gemini", "groq", "openai", "ollama", "mock"}, func(s string) {
		switch s {
		case "gemini":
			v.modelEntry.SetText("gemini-2.0-flash")
			v.baseURLEntry.SetText("")
		case "groq":
			v.modelEntry.SetText("llama-3.3-70b-versatile")
			v.baseURLEntry.SetText("https://api.groq.com/openai/v1")
		case "openai":
			v.modelEntry.SetText("gpt-4o")
			v.baseURLEntry.SetText("https://api.openai.com/v1")
		case "ollama":
			v.modelEntry.SetText("llama3")
			v.baseURLEntry.SetText("http://localhost:11434")
		case "mock":
			v.modelEntry.SetText("mock-engine")
			v.baseURLEntry.SetText("")
		}
	})

	v.apiKeyEntry = widget.NewPasswordEntry()
	v.apiKeyEntry.SetPlaceHolder("Clave API (se cifrará con AES-256 GCM)...")

	v.modelEntry = widget.NewEntry()
	v.modelEntry.SetPlaceHolder("Modelo de IA predeterminado...")

	v.baseURLEntry = widget.NewEntry()
	v.baseURLEntry.SetPlaceHolder("URL Base (opcional o requerido para Ollama/Groq)...")

	v.activeCheck = widget.NewCheck("Establecer como Proveedor de IA Activo", nil)
	v.activeCheck.SetChecked(true)

	v.statusLabel = widget.NewLabelWithStyle("Proveedor Activo Actual: Verificando...", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	btnSave := widget.NewButtonWithIcon("Guardar Configuración Cifrada", theme.DocumentSaveIcon(), func() {
		if v.providerSelect.Selected == "" {
			dialog.ShowInformation("Atención", "Seleccione un proveedor de IA", v.win)
			return
		}
		err := v.configService.SaveProviderConfig(
			context.Background(),
			v.providerSelect.Selected,
			v.apiKeyEntry.Text,
			v.modelEntry.Text,
			v.baseURLEntry.Text,
			v.activeCheck.Checked,
		)
		if err != nil {
			dialog.ShowError(err, v.win)
			return
		}
		v.reloadConfigs()
		v.apiKeyEntry.SetText("")
		dialog.ShowInformation("Éxito", "Configuración guardada y cifrada correctamente con AES-256 GCM", v.win)
	})
	btnSave.Importance = widget.HighImportance

	settingsForm := widget.NewForm(
		widget.NewFormItem("Proveedor de IA:", v.providerSelect),
		widget.NewFormItem("Clave API (API Key):", v.apiKeyEntry),
		widget.NewFormItem("Modelo Predeterminado:", v.modelEntry),
		widget.NewFormItem("URL Base (Opcional / Ollama):", v.baseURLEntry),
	)

	formCard := widget.NewCard("Configurar Proveedor de IA", "Las claves API se almacenan localmente y cifradas con AES-256 GCM.", container.NewVBox(
		v.statusLabel,
		settingsForm,
		v.activeCheck,
		btnSave,
	))

	v.configList = widget.NewList(
		func() int { return len(v.configs) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel("proveedor"), widget.NewLabel("modelo"), widget.NewLabel("estado"))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(v.configs) {
				c := v.configs[id]
				active := "Inactivo"
				if c.IsActive {
					active = "★ ACTIVO"
				}
				box := obj.(*fyne.Container)
				box.Objects[0].(*widget.Label).SetText(c.ProviderName)
				box.Objects[1].(*widget.Label).SetText(fmt.Sprintf("(%s)", c.DefaultModel))
				box.Objects[2].(*widget.Label).SetText(active)
			}
		},
	)
	v.configList.OnSelected = func(id widget.ListItemID) {
		if id < len(v.configs) {
			c := v.configs[id]
			v.providerSelect.SetSelected(c.ProviderName)
			v.modelEntry.SetText(c.DefaultModel)
			v.baseURLEntry.SetText(c.BaseURL)
			v.activeCheck.SetChecked(c.IsActive)
		}
	}

	btnActivate := widget.NewButton("Activar Proveedor Seleccionado", func() {
		if v.providerSelect.Selected == "" {
			return
		}
		_ = v.configService.SetActiveProvider(context.Background(), v.providerSelect.Selected)
		v.reloadConfigs()
	})

	listCard := widget.NewCard("Proveedores Registrados", "", container.NewBorder(
		nil,
		btnActivate,
		nil, nil,
		v.configList,
	))

	split := container.NewHSplit(formCard, listCard)
	split.SetOffset(0.55)

	v.providerSelect.SetSelected("gemini")
	v.reloadConfigs()
	return split
}

func (v *SettingsView) reloadConfigs() {
	cfgs, err := v.configService.ListConfigs(context.Background())
	if err == nil {
		v.configs = cfgs
		v.configList.Refresh()
	}

	active, err := v.configService.GetActiveConfig(context.Background())
	if err == nil && active != nil {
		v.statusLabel.SetText(fmt.Sprintf("Proveedor Activo Actual: %s (%s)", active.ProviderName, active.DefaultModel))
	} else {
		v.statusLabel.SetText("Proveedor Activo Actual: Modo Mock Local")
	}
}
