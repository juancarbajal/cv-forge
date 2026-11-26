package filestore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type MarkdownStore struct {
	baseDir string
}

func NewMarkdownStore(baseDir string) (*MarkdownStore, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio de CVs: %w", err)
	}
	return &MarkdownStore{baseDir: baseDir}, nil
}

// SaveContent guarda el contenido de un CV en formato markdown: {baseDir}/{cvID}_v{version}.md
func (s *MarkdownStore) SaveContent(ctx context.Context, cvID string, version int, content string) (string, error) {
	filename := fmt.Sprintf("%s_v%d.md", cvID, version)
	fullPath := filepath.Join(s.baseDir, filename)

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("error escribiendo archivo markdown: %w", err)
	}

	return fullPath, nil
}

// GetContent lee el contenido del archivo markdown en la ruta especificada
func (s *MarkdownStore) GetContent(ctx context.Context, filePath string) (string, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error leyendo archivo markdown: %w", err)
	}
	return string(bytes), nil
}

// DeleteContent elimina el archivo físico si existe
func (s *MarkdownStore) DeleteContent(ctx context.Context, filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(filePath)
}
