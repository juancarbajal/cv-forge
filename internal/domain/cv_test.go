package domain

import (
	"testing"
	"time"
)

func TestCV_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cv      CV
		wantErr bool
	}{
		{
			name: "CV válido",
			cv: CV{
				ID:              "cv-1",
				Title:           "CV Backend",
				PrimaryLanguage: LangSpanish,
			},
			wantErr: false,
		},
		{
			name: "Falta ID",
			cv: CV{
				Title:           "CV Backend",
				PrimaryLanguage: LangSpanish,
			},
			wantErr: true,
		},
		{
			name: "Falta Título",
			cv: CV{
				ID:              "cv-1",
				Title:           "   ",
				PrimaryLanguage: LangSpanish,
			},
			wantErr: true,
		},
		{
			name: "Asigna idioma español por defecto si está vacío",
			cv: CV{
				ID:    "cv-1",
				Title: "CV Backend",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cv.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CV.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.cv.PrimaryLanguage == "" {
				t.Errorf("expected default language assigned, got empty")
			}
		})
	}
}

func TestCVVersion_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ver     CVVersion
		wantErr bool
	}{
		{
			name: "Versión válida",
			ver: CVVersion{
				ID:            "ver-1",
				CVID:          "cv-1",
				VersionNumber: 1,
				FilePath:      "data/cvs/cv-1_v1.md",
				CreatedAt:     time.Now(),
			},
			wantErr: false,
		},
		{
			name: "Número de versión inválido <= 0",
			ver: CVVersion{
				ID:            "ver-1",
				CVID:          "cv-1",
				VersionNumber: 0,
				FilePath:      "data/cvs/cv-1_v1.md",
			},
			wantErr: true,
		},
		{
			name: "Falta FilePath",
			ver: CVVersion{
				ID:            "ver-1",
				CVID:          "cv-1",
				VersionNumber: 1,
				FilePath:      "",
			},
			wantErr: true,
		},
		{
			name: "Falta CVID",
			ver: CVVersion{
				ID:            "ver-1",
				CVID:          "",
				VersionNumber: 1,
				FilePath:      "data/cvs/cv-1_v1.md",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ver.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CVVersion.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
