package domain

import (
	"testing"
)

func TestAntiInventionValidator_Success(t *testing.T) {
	validator := NewAntiInventionValidator()

	originalCV := `# Juan Pérez
Experiencia como Desarrollador Backend utilizando bases de datos relacionales y contenedores Linux.
`

	optResult := &OptimizedCVResult{
		Changes: []OptimizationChange{
			{
				ChangeSummary: "Se enfatizó PostgreSQL en servicios backend",
				EvidenceUsed:  "mencionaba bases de datos relacionales",
			},
		},
	}

	alerts, err := validator.Validate(originalCV, optResult)
	if err != nil {
		t.Fatalf("expected validation success, got error: %v", err)
	}

	for _, a := range alerts {
		if a.Level == "BLOCKING" {
			t.Errorf("unexpected blocking alert: %v", a)
		}
	}
}

func TestAntiInventionValidator_BlockingEmptyEvidence(t *testing.T) {
	validator := NewAntiInventionValidator()

	originalCV := `# Juan Pérez`
	optResult := &OptimizedCVResult{
		Changes: []OptimizationChange{
			{
				ChangeSummary: "Se agregó maestría en Harvard",
				EvidenceUsed:  "", // Sin evidencia
			},
		},
	}

	_, err := validator.Validate(originalCV, optResult)
	if err == nil {
		t.Fatalf("expected error for empty evidence, got nil")
	}
}
