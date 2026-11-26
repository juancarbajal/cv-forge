package filestore

import (
	"context"
	"os"
	"testing"
)

func TestMarkdownStore(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "md_store_test_*")
	if err != nil {
		t.Fatalf("error creando tempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := NewMarkdownStore(tempDir)
	if err != nil {
		t.Fatalf("error inicializando store: %v", err)
	}

	ctx := context.Background()
	content := "# Mi CV\nExperiencia en Go."

	path, err := store.SaveContent(ctx, "cv-1", 1, content)
	if err != nil {
		t.Fatalf("error guardando contenido: %v", err)
	}

	readContent, err := store.GetContent(ctx, path)
	if err != nil {
		t.Fatalf("error leyendo contenido: %v", err)
	}

	if readContent != content {
		t.Fatalf("contenido leído no coincide: '%s' vs '%s'", readContent, content)
	}

	if err := store.DeleteContent(ctx, path); err != nil {
		t.Fatalf("error eliminando contenido: %v", err)
	}

	_, err = store.GetContent(ctx, path)
	if err == nil {
		t.Fatalf("esperaba error al leer archivo eliminado")
	}
}
