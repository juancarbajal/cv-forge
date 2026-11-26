package views

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/jcarbajal/cv-checker/internal/app"
	"github.com/jcarbajal/cv-checker/internal/domain"
)

type AuditView struct {
	auditService *app.AuditService
	win          fyne.Window

	logList *widget.List
	logs    []domain.AuditLog
}

func NewAuditView(auditService *app.AuditService, win fyne.Window) fyne.CanvasObject {
	v := &AuditView{
		auditService: auditService,
		win:          win,
	}
	return v.buildUI()
}

func (v *AuditView) buildUI() fyne.CanvasObject {
	v.logList = widget.NewList(
		func() int { return len(v.logs) },
		func() fyne.CanvasObject {
			return container.NewVBox(
				container.NewHBox(
					widget.NewLabelWithStyle("Timestamp", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
					widget.NewLabel("Acción"),
					widget.NewLabel("Proveedor (Modelo)"),
					widget.NewLabel("Resultado"),
				),
				widget.NewLabel("Detalles"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(v.logs) {
				l := v.logs[id]
				box := obj.(*fyne.Container)
				topRow := box.Objects[0].(*fyne.Container)

				topRow.Objects[0].(*widget.Label).SetText(l.Timestamp.Format("2006-01-02 15:04:05 UTC"))
				topRow.Objects[1].(*widget.Label).SetText(string(l.Action))
				prov := l.Provider
				if l.Model != "" {
					prov = fmt.Sprintf("%s (%s)", prov, l.Model)
				}
				topRow.Objects[2].(*widget.Label).SetText(prov)
				topRow.Objects[3].(*widget.Label).SetText(l.ResultStatus)

				box.Objects[1].(*widget.Label).SetText(l.DetailsJSON)
			}
		},
	)

	btnRefresh := widget.NewButton("Actualizar Registro de Auditoría", func() {
		v.reloadLogs()
	})

	card := widget.NewCard("Historial de Auditoría Inmutable", "Registro de todas las operaciones críticas, llamadas a IA y optimizaciones.", container.NewBorder(
		nil,
		btnRefresh,
		nil, nil,
		v.logList,
	))

	v.reloadLogs()
	return card
}

func (v *AuditView) reloadLogs() {
	logs, err := v.auditService.List(context.Background(), 100)
	if err == nil {
		v.logs = logs
		v.logList.Refresh()
	}
}
