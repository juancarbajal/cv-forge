package domain

import (
	"fmt"
	"strings"
)

type ValidationAlert struct {
	Level       string `json:"level"` // "WARNING", "BLOCKING"
	Description string `json:"description"`
	Item        string `json:"item"`
}

type AntiInventionValidator struct{}

func NewAntiInventionValidator() *AntiInventionValidator {
	return &AntiInventionValidator{}
}

// Validate verifica que las mejoras propuestas en el CV optimizado cuenten con evidencia sustentable en el CV origen
func (v *AntiInventionValidator) Validate(originalCV string, optResult *OptimizedCVResult) ([]ValidationAlert, error) {
	var alerts []ValidationAlert
	origLower := strings.ToLower(originalCV)

	for _, change := range optResult.Changes {
		evidence := strings.TrimSpace(change.EvidenceUsed)
		summary := strings.TrimSpace(change.ChangeSummary)

		// 1. Verificar que todo cambio declare explícitamente su evidencia de origen
		if evidence == "" {
			alerts = append(alerts, ValidationAlert{
				Level:       "BLOCKING",
				Description: "Cambio sin evidencia de origen declarada",
				Item:        summary,
			})
			continue
		}

		// 2. Verificar que las palabras clave de la evidencia existan o tengan sentido semántico en el CV origen
		evidenceWords := strings.Fields(strings.ToLower(evidence))
		matchCount := 0
		significantWords := 0

		for _, word := range evidenceWords {
			// Ignorar palabras de enlace cortas
			if len(word) <= 3 {
				continue
			}
			significantWords++
			if strings.Contains(origLower, word) {
				matchCount++
			}
		}

		if significantWords > 0 && matchCount == 0 {
			alerts = append(alerts, ValidationAlert{
				Level:       "WARNING",
				Description: fmt.Sprintf("La evidencia citada ('%s') no parece estar presente textualmente en el CV de origen", evidence),
				Item:        summary,
			})
		}
	}

	for _, alert := range alerts {
		if alert.Level == "BLOCKING" {
			return alerts, fmt.Errorf("%w: %s (%s)", ErrAntiInventionFailed, alert.Description, alert.Item)
		}
	}

	return alerts, nil
}
