package domain

import (
	"testing"
)

func TestJobApplication_Validate(t *testing.T) {
	tests := []struct {
		name    string
		job     JobApplication
		wantErr bool
	}{
		{
			name: "Job válido",
			job: JobApplication{
				ID:            "job-1",
				Company:       "Globant",
				Role:          "Go Developer",
				JobPostingRaw: "Buscamos Go Developer con experiencia...",
				CVID:          "cv-1",
			},
			wantErr: false,
		},
		{
			name: "Falta ID",
			job: JobApplication{
				Company:       "Globant",
				Role:          "Go Developer",
				JobPostingRaw: "Buscamos Go Developer...",
				CVID:          "cv-1",
			},
			wantErr: true,
		},
		{
			name: "Falta Empresa",
			job: JobApplication{
				ID:            "job-1",
				Company:       "",
				Role:          "Go Developer",
				JobPostingRaw: "Buscamos Go Developer...",
				CVID:          "cv-1",
			},
			wantErr: true,
		},
		{
			name: "Falta Cargo",
			job: JobApplication{
				ID:            "job-1",
				Company:       "Globant",
				Role:          "",
				JobPostingRaw: "Buscamos Go Developer...",
				CVID:          "cv-1",
			},
			wantErr: true,
		},
		{
			name: "Falta Texto de la Oferta",
			job: JobApplication{
				ID:            "job-1",
				Company:       "Globant",
				Role:          "Go Developer",
				JobPostingRaw: "",
				CVID:          "cv-1",
			},
			wantErr: true,
		},
		{
			name: "Falta CV seleccionado",
			job: JobApplication{
				ID:            "job-1",
				Company:       "Globant",
				Role:          "Go Developer",
				JobPostingRaw: "Buscamos Go Developer...",
				CVID:          "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.job.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("JobApplication.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
