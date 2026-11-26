package views

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// Paleta semántica compartida para badges y estados en todas las vistas.
var (
	ColorSuccess = color.NRGBA{R: 22, G: 163, B: 74, A: 255}
	ColorWarning = color.NRGBA{R: 217, G: 119, B: 6, A: 255}
	ColorDanger  = color.NRGBA{R: 220, G: 38, B: 38, A: 255}
	ColorInfo    = color.NRGBA{R: 37, G: 99, B: 235, A: 255}
	ColorNeutral = color.NRGBA{R: 100, G: 116, B: 139, A: 255}
)

// MakeStatusBadge crea una insignia visual redondeada con color semántico.
func MakeStatusBadge(label string, bg color.NRGBA) fyne.CanvasObject {
	rect := canvas.NewRectangle(bg)
	rect.CornerRadius = 6
	txt := canvas.NewText(label, color.White)
	txt.TextSize = 11
	txt.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewStack(rect, container.NewPadded(txt))
}
