package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// AppTheme envuelve el tema por defecto de Fyne para refinar espaciado, radios y acentos de cv-checker.
type AppTheme struct {
	fyne.Theme
}

// NewAppTheme devuelve el tema personalizado de la aplicación.
func NewAppTheme() fyne.Theme {
	return &AppTheme{Theme: theme.DefaultTheme()}
}

func (t *AppTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 37, G: 99, B: 235, A: 255} // Azul profesional (#2563EB)
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 22, G: 163, B: 74, A: 255}
	case theme.ColorNameWarning:
		return color.NRGBA{R: 217, G: 119, B: 6, A: 255}
	case theme.ColorNameError:
		return color.NRGBA{R: 220, G: 38, B: 38, A: 255}
	default:
		return t.Theme.Color(name, variant)
	}
}

func (t *AppTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameInputRadius:
		return 6
	default:
		return t.Theme.Size(name)
	}
}
