package views

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// RunAsync ejecuta una función pesada en segundo plano con timeout, mostrando un diálogo de progreso cancelable
func RunAsync(win fyne.Window, title, message string, timeout time.Duration, task func(ctx context.Context) error, onComplete func(err error)) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	progressDialog := dialog.NewCustomWithoutButtons(title, nil, win)
	progressDialog.Show()

	go func() {
		defer cancel()
		err := task(ctx)

		fyne.Do(func() {
			progressDialog.Hide()
			if onComplete != nil {
				onComplete(err)
			}
		})
	}()
}
