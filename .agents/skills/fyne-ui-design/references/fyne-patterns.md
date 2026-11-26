# Fyne v2 Patterns Reference (`cv-checker`)

## 1. Primary vs. Secondary Button Importance
```go
primaryBtn := widget.NewButtonWithIcon("Optimizar CV", theme.ConfirmIcon(), onOptimize)
primaryBtn.Importance = widget.HighImportance

secondaryBtn := widget.NewButtonWithIcon("Responder Gaps", theme.QuestionIcon(), onGapInterview)
secondaryBtn.Importance = widget.MediumImportance
```

## 2. Semantic Status Badges (ATS Matches / Agent Verdicts)
Use `canvas.NewRectangle` inside a `container.NewStack` with padded `widget.Label` or `canvas.NewText` for crisp status pills:
```go
func makeStatusBadge(label string, bg color.NRGBA) fyne.CanvasObject {
	rect := canvas.NewRectangle(bg)
	rect.CornerRadius = 6
	txt := canvas.NewText(label, color.White)
	txt.TextSize = 11
	txt.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewStack(rect, container.NewPadded(txt))
}
```
Recommended semantic palette:
- **Cumple / Approved:** `color.NRGBA{R: 22, G: 163, B: 74, A: 255}`
- **Cumple parcialmente / Warning:** `color.NRGBA{R: 217, G: 119, B: 6, A: 255}`
- **No evidenciado / Blocking:** `color.NRGBA{R: 220, G: 38, B: 38, A: 255}`

## 3. Custom `fyne.Theme` Skeleton
Wrap `theme.DefaultTheme()` so only the tokens you customize are overridden:
```go
type AppTheme struct{ fyne.Theme }

func NewAppTheme() fyne.Theme {
	return &AppTheme{Theme: theme.DefaultTheme()}
}

func (t *AppTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 10
	default:
		return t.Theme.Size(name)
	}
}
```

## 4. Thread-Safe Async UI Updates
Always wrap UI state mutations inside `fyne.Do` when invoked from background goroutines:
```go
fyne.Do(func() {
	label.SetText("Actualizado")
	container.Refresh()
})
```

## 5. Headless Widget Testing (`fyne.io/fyne/v2/test`)
```go
func TestMyWidget(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := test.NewWindow(myCanvasObj)
	test.Tap(myButton)
	// Assert label text or state
}
```
