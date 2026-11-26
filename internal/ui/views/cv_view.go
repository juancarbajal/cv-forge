package views

import (
	"context"
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/jcarbajal/cv-checker/internal/app"
	"github.com/jcarbajal/cv-checker/internal/domain"
)

type CVView struct {
	cvService *app.CVService
	win       fyne.Window

	cvList       *widget.List
	cvs          []domain.CV
	selectedCV   *domain.CV
	selectedVer  *domain.CVVersion
	versionList  *widget.List

	titleLabel   *widget.Label
	descLabel    *widget.Label
	langLabel    *widget.Label
	favLabel     *widget.Label
	editorEntry  *widget.Entry
	reasonEntry  *widget.Entry
}

func NewCVView(cvService *app.CVService, win fyne.Window) fyne.CanvasObject {
	v := &CVView{
		cvService: cvService,
		win:       win,
	}
	return v.buildUI()
}

func (v *CVView) buildUI() fyne.CanvasObject {
	// 1. Panel Izquierdo: Lista de CVs y Acciones
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Buscar CV...")
	searchEntry.OnChanged = func(s string) {
		v.reloadCVs(s)
	}

	v.cvList = widget.NewList(
		func() int { return len(v.cvs) },
		func() fyne.CanvasObject {
			return widget.NewLabel("CV Template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(v.cvs) {
				c := v.cvs[id]
				fav := ""
				if c.FavoriteVersionID != "" {
					fav = " ★"
				}
				obj.(*widget.Label).SetText(fmt.Sprintf("%s (%s)%s", c.Title, c.PrimaryLanguage, fav))
			}
		},
	)
	v.cvList.OnSelected = func(id widget.ListItemID) {
		if id < len(v.cvs) {
			v.selectCV(v.cvs[id].ID)
		}
	}

	btnNew := widget.NewButton("Crear Nuevo CV", func() {
		v.showCreateDialog(false)
	})
	btnImport := widget.NewButton("Importar Markdown", func() {
		v.showCreateDialog(true)
	})
	btnDuplicate := widget.NewButton("Duplicar", func() {
		if v.selectedCV == nil {
			dialog.ShowInformation("Atención", "Seleccione un CV para duplicar", v.win)
			return
		}
		newTitle := fmt.Sprintf("%s (Copia)", v.selectedCV.Title)
		_, err := v.cvService.DuplicateCV(context.Background(), v.selectedCV.ID, newTitle)
		if err != nil {
			dialog.ShowError(err, v.win)
			return
		}
		v.reloadCVs("")
	})
	btnDelete := widget.NewButton("Eliminar", func() {
		if v.selectedCV == nil {
			return
		}
		dialog.ShowConfirm("Confirmar", fmt.Sprintf("¿Eliminar el CV '%s' y todas sus versiones?", v.selectedCV.Title), func(ok bool) {
			if ok {
				_ = v.cvService.DeleteCV(context.Background(), v.selectedCV.ID)
				v.selectedCV = nil
				v.reloadCVs("")
				v.clearDetails()
			}
		}, v.win)
	})

	leftButtons := container.NewGridWithColumns(2, btnNew, btnImport, btnDuplicate, btnDelete)
	leftPanel := container.NewBorder(
		container.NewVBox(widget.NewLabelWithStyle("Gestor de CVs", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), searchEntry),
		leftButtons,
		nil,
		nil,
		v.cvList,
	)

	// 2. Panel Derecho: Detalles del CV, Versiones y Editor
	v.titleLabel = widget.NewLabelWithStyle("Seleccione un CV", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	v.descLabel = widget.NewLabel("")
	v.langLabel = widget.NewLabel("")
	v.favLabel = widget.NewLabel("")

	metaCard := widget.NewCard("Detalles del CV", "", container.NewVBox(
		v.titleLabel,
		v.descLabel,
		container.NewHBox(v.langLabel, v.favLabel),
	))

	// Lista de versiones
	v.versionList = widget.NewList(
		func() int {
			if v.selectedCV == nil {
				return 0
			}
			return len(v.selectedCV.Versions)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel("v1"), widget.NewLabel("motivo"))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if v.selectedCV != nil && id < len(v.selectedCV.Versions) {
				ver := v.selectedCV.Versions[id]
				fav := ""
				if ver.IsFavorite {
					fav = " ★ Favorita"
				}
				box := obj.(*fyne.Container)
				box.Objects[0].(*widget.Label).SetText(fmt.Sprintf("v%d%s", ver.VersionNumber, fav))
				box.Objects[1].(*widget.Label).SetText(fmt.Sprintf("- %s", ver.ChangeReason))
			}
		},
	)
	v.versionList.OnSelected = func(id widget.ListItemID) {
		if v.selectedCV != nil && id < len(v.selectedCV.Versions) {
			v.selectVersion(&v.selectedCV.Versions[id])
		}
	}

	btnSetFav := widget.NewButton("Marcar Favorita (★)", func() {
		if v.selectedCV == nil || v.selectedVer == nil {
			dialog.ShowInformation("Atención", "Seleccione una versión para marcar como favorita", v.win)
			return
		}
		err := v.cvService.SetFavoriteVersion(context.Background(), v.selectedCV.ID, v.selectedVer.ID)
		if err != nil {
			dialog.ShowError(err, v.win)
			return
		}
		v.selectCV(v.selectedCV.ID)
	})

	versionBox := container.NewBorder(
		widget.NewLabelWithStyle("Historial de Versiones", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		btnSetFav,
		nil,
		nil,
		v.versionList,
	)

	// Editor Markdown de la versión
	v.editorEntry = widget.NewMultiLineEntry()
	v.editorEntry.SetPlaceHolder("Escriba o pegue el contenido de su CV en formato Markdown...")

	v.reasonEntry = widget.NewEntry()
	v.reasonEntry.SetPlaceHolder("Ej: Se incorporaron certificaciones en AWS y proyectos recientes con Go...")

	btnSaveVersion := widget.NewButton("Guardar como Nueva Versión", func() {
		if v.selectedCV == nil {
			return
		}
		reason := v.reasonEntry.Text
		if reason == "" {
			reason = "Edición manual de usuario"
		}
		srcID := ""
		if v.selectedVer != nil {
			srcID = v.selectedVer.ID
		}
		_, err := v.cvService.AddNewVersion(context.Background(), v.selectedCV.ID, srcID, v.editorEntry.Text, reason, "Manual", "Editor")
		if err != nil {
			dialog.ShowError(err, v.win)
			return
		}
		v.reasonEntry.SetText("")
		v.selectCV(v.selectedCV.ID)
		dialog.ShowInformation("Éxito", "Nueva versión guardada correctamente", v.win)
	})

	reasonForm := widget.NewForm(
		widget.NewFormItem("Motivo de esta Versión:", v.reasonEntry),
	)

	editorContainer := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Editor de Contenido Markdown", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			reasonForm,
			widget.NewLabelWithStyle("Contenido del CV (Markdown):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		),
		btnSaveVersion,
		nil,
		nil,
		v.editorEntry,
	)

	rightSplit := container.NewHSplit(versionBox, editorContainer)
	rightSplit.SetOffset(0.35)

	rightPanel := container.NewBorder(metaCard, nil, nil, nil, rightSplit)

	mainSplit := container.NewHSplit(leftPanel, rightPanel)
	mainSplit.SetOffset(0.3)

	v.reloadCVs("")
	return mainSplit
}

func (v *CVView) reloadCVs(search string) {
	cvs, err := v.cvService.ListCVs(context.Background(), search)
	if err == nil {
		v.cvs = cvs
		v.cvList.Refresh()
	}
}

func (v *CVView) selectCV(id string) {
	cv, err := v.cvService.GetCV(context.Background(), id)
	if err != nil {
		return
	}
	v.selectedCV = cv
	v.titleLabel.SetText(cv.Title)
	v.descLabel.SetText(cv.Description)
	v.langLabel.SetText(fmt.Sprintf("Idioma: %s", cv.PrimaryLanguage))
	v.favLabel.SetText(fmt.Sprintf("Versión Favorita: %s", cv.FavoriteVersionID))
	v.versionList.Refresh()

	// Seleccionar versión favorita o la última por defecto
	var targetVer *domain.CVVersion
	if cv.FavoriteVersionID != "" {
		for i := range cv.Versions {
			if cv.Versions[i].ID == cv.FavoriteVersionID {
				targetVer = &cv.Versions[i]
				break
			}
		}
	}
	if targetVer == nil && len(cv.Versions) > 0 {
		targetVer = &cv.Versions[len(cv.Versions)-1]
	}
	if targetVer != nil {
		v.selectVersion(targetVer)
	}
}

func (v *CVView) selectVersion(ver *domain.CVVersion) {
	v.selectedVer = ver
	content, _, err := v.cvService.GetVersionContent(context.Background(), ver.ID)
	if err == nil {
		v.editorEntry.SetText(content)
	}
}

func (v *CVView) clearDetails() {
	v.titleLabel.SetText("Seleccione un CV")
	v.descLabel.SetText("")
	v.langLabel.SetText("")
	v.favLabel.SetText("")
	v.editorEntry.SetText("")
	v.versionList.Refresh()
}

func (v *CVView) showCreateDialog(isImport bool) {
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Ej: CV Backend Go Senior...")

	descEntry := widget.NewEntry()
	descEntry.SetPlaceHolder("Ej: Especialización en microservicios y bases de datos relacionales...")

	langSelect := widget.NewSelect([]string{"Español (es)", "Inglés (en)", "Portugués (pt)"}, nil)
	langSelect.SetSelected("Español (es)")

	contentEntry := widget.NewMultiLineEntry()
	contentEntry.SetPlaceHolder("# Tu Nombre Completo\n\n## Resumen Profesional\nExperiencia en desarrollo backend...")

	var formItems []*widget.FormItem
	formItems = append(formItems, widget.NewFormItem("Nombre del CV:", titleEntry))
	formItems = append(formItems, widget.NewFormItem("Descripción / Perfil:", descEntry))
	formItems = append(formItems, widget.NewFormItem("Idioma Principal:", langSelect))

	if isImport {
		btnFile := widget.NewButton("Seleccionar archivo Markdown (.md)", func() {
			fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil || reader == nil {
					return
				}
				defer reader.Close()
				bytes, err := os.ReadFile(reader.URI().Path())
				if err == nil {
					contentEntry.SetText(string(bytes))
					if titleEntry.Text == "" {
						titleEntry.SetText(reader.URI().Name())
					}
				}
			}, v.win)
			fileDialog.Show()
		})
		formItems = append(formItems, widget.NewFormItem("Archivo a Importar:", btnFile))
	}
	formItems = append(formItems, widget.NewFormItem("Contenido Markdown:", contentEntry))

	dialog.ShowForm("Crear Nuevo CV", "Guardar", "Cancelar", formItems, func(ok bool) {
		if !ok || titleEntry.Text == "" {
			return
		}
		var lang domain.Language = domain.LangSpanish
		if langSelect.Selected == "Inglés (en)" {
			lang = domain.LangEnglish
		} else if langSelect.Selected == "Portugués (pt)" {
			lang = domain.LangPortuguese
		}

		content := contentEntry.Text
		if content == "" {
			content = fmt.Sprintf("# %s\n\nResumen profesional...", titleEntry.Text)
		}

		cv, _, err := v.cvService.CreateCV(context.Background(), titleEntry.Text, descEntry.Text, lang, content)
		if err != nil {
			dialog.ShowError(err, v.win)
			return
		}
		v.reloadCVs("")
		v.selectCV(cv.ID)
	}, v.win)
}
