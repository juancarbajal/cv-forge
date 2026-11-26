---
name: fyne-ui-design
description: >-
  Guides UI/UX improvements, custom theming, layout hierarchy, and thread-safe
  widget patterns for Go + Fyne v2 desktop views in cv-checker. Use when
  auditing, redesigning, or polishing files under internal/ui/.
---

# Fyne v2 UI/UX Design Skill

Use this skill when improving the visual quality, usability, responsiveness, or layout structure of `cv-checker`'s Fyne v2 desktop interface (`internal/ui/`).

## Workflow

### 1. Audit the Target View
- Read the target view in `internal/ui/views/` and `internal/ui/app.go`.
- Check for:
  - Visual hierarchy (missing `widget.Card` grouping, flat button rows, unclear primary vs. secondary actions).
  - Overflow or cramped panels (missing `container.NewVScroll`, unbalanced `HSplit` offsets, unwrapped long labels).
  - Thread safety: all UI mutations from goroutines or `RunAsync` callbacks must run inside `fyne.Do(func() { ... })`.
  - Separation of concerns: keep business logic in `internal/app/` and domain models in `internal/domain/`.

### 2. Apply Visual & Layout Improvements
- **Theme & Tokens:** Prefer a centralized `fyne.Theme` implementation in `internal/ui/theme.go` (or `internal/ui/theme/`) for consistent padding, corner radius, typography scale, and semantic colors (success, warning, danger, muted).
- **Hierarchy:** Highlight the primary CTA with `widget.HighImportance` (`Importance = widget.HighImportance`) and keep secondary/optional actions at medium/low importance.
- **Scannability:** Replace dense raw text blocks with structured cards, status badges, and clear section headers.
- **Responsiveness:** Keep `RunAsync` timeouts aligned with backend work (e.g., `150s` for multi-step agentic loops) and show clear loading/error states.

### 3. Consult Patterns Reference
- Read [fyne-patterns.md](./references/fyne-patterns.md) when implementing custom `fyne.Theme`, status badges, or headless Fyne widget tests.

### 4. Verify
- Run `go vet ./internal/ui/...` or `make build` (when Fyne/GLFW system headers are available).
- Run `go test ./internal/ui/...` using `fyne.io/fyne/v2/test` for widget behavior.
- Confirm backend tests still pass: `go test ./internal/domain/... ./internal/app/... ./internal/infrastructure/...`.
