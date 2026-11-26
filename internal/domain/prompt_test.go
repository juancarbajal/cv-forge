package domain

import (
	"testing"
)

func TestPrompt_Validate(t *testing.T) {
	tests := []struct {
		name    string
		prompt  Prompt
		wantErr bool
	}{
		{
			name: "Prompt válido",
			prompt: Prompt{
				ID:      "prompt-1",
				Type:    PromptTypeJobAnalysis,
				Name:    "Análisis Base",
				Version: 1,
				Content: "Extrae requisitos técnicos...",
			},
			wantErr: false,
		},
		{
			name: "Falta ID",
			prompt: Prompt{
				Type:    PromptTypeJobAnalysis,
				Name:    "Análisis Base",
				Version: 1,
				Content: "Extrae requisitos...",
			},
			wantErr: true,
		},
		{
			name: "Falta Tipo",
			prompt: Prompt{
				ID:      "prompt-1",
				Type:    "",
				Name:    "Análisis Base",
				Version: 1,
				Content: "Extrae requisitos...",
			},
			wantErr: true,
		},
		{
			name: "Versión <= 0",
			prompt: Prompt{
				ID:      "prompt-1",
				Type:    PromptTypeJobAnalysis,
				Name:    "Análisis Base",
				Version: 0,
				Content: "Extrae requisitos...",
			},
			wantErr: true,
		},
		{
			name: "Falta Contenido",
			prompt: Prompt{
				ID:      "prompt-1",
				Type:    PromptTypeJobAnalysis,
				Name:    "Análisis Base",
				Version: 1,
				Content: "   ",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.prompt.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Prompt.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
