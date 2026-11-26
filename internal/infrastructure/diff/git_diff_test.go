package diff

import (
	"strings"
	"testing"
)

func TestGitDiffEngine_Modifications(t *testing.T) {
	orig := `Line 1
Experience developing backend services.
Line 3`

	mod := `Line 1
Experience developing backend services using Go and PostgreSQL.
Line 3`

	engine := NewGitDiffEngine()
	diffStr, err := engine.GenerateDiff(orig, mod)
	if err != nil {
		t.Fatalf("error generando diff: %v", err)
	}

	if !strings.Contains(diffStr, "- Experience developing backend services.") {
		t.Errorf("diff no contiene línea eliminada esperada: %s", diffStr)
	}
	if !strings.Contains(diffStr, "+ Experience developing backend services using Go and PostgreSQL.") {
		t.Errorf("diff no contiene línea agregada esperada: %s", diffStr)
	}
	if !strings.Contains(diffStr, "  Line 1") {
		t.Errorf("diff no contiene línea de contexto esperada: %s", diffStr)
	}
}

func TestGitDiffEngine_Identical(t *testing.T) {
	text := "Línea 1\nLínea 2\nLínea 3\n"
	engine := NewGitDiffEngine()
	diffStr, err := engine.GenerateDiff(text, text)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	for _, line := range strings.Split(diffStr, "\n") {
		if strings.HasPrefix(line, "+ ") || strings.HasPrefix(line, "- ") {
			t.Errorf("no deberían haber cambios en diff idéntico: %s", line)
		}
	}
}

func TestGitDiffEngine_EmptyOriginal(t *testing.T) {
	engine := NewGitDiffEngine()
	diffStr, err := engine.GenerateDiff("", "Línea nueva\nSegunda línea\n")
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if !strings.Contains(diffStr, "+ Línea nueva") {
		t.Errorf("se esperaba adición: %s", diffStr)
	}
}

func TestGitDiffEngine_CRLFNormalization(t *testing.T) {
	orig := "Line 1\r\nLine 2\r\n"
	mod := "Line 1\nLine 2 modified\n"

	engine := NewGitDiffEngine()
	diffStr, err := engine.GenerateDiff(orig, mod)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if !strings.Contains(diffStr, "- Line 2") || !strings.Contains(diffStr, "+ Line 2 modified") {
		t.Errorf("normalización CRLF falló: %s", diffStr)
	}
}
