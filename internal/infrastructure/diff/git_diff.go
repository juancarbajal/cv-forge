package diff

import (
	"fmt"
	"strings"
)

type GitDiffEngine struct{}

func NewGitDiffEngine() *GitDiffEngine {
	return &GitDiffEngine{}
}

// GenerateDiff genera una salida de diff estilo Git (formato unificado) entre dos textos
func (g *GitDiffEngine) GenerateDiff(original, modified string) (string, error) {
	origLines := splitLines(original)
	modLines := splitLines(modified)

	lcs := computeLCS(origLines, modLines)

	var sb strings.Builder
	sb.WriteString("--- Original\n+++ Optimizado\n")

	i, j := 0, 0
	for i < len(origLines) || j < len(modLines) {
		if i < len(origLines) && j < len(modLines) && origLines[i] == modLines[j] {
			sb.WriteString(fmt.Sprintf("  %s\n", origLines[i]))
			i++
			j++
		} else if j < len(modLines) && (i >= len(origLines) || lcs[i][j+1] >= lcs[i+1][j]) {
			sb.WriteString(fmt.Sprintf("+ %s\n", modLines[j]))
			j++
		} else if i < len(origLines) && (j >= len(modLines) || lcs[i][j+1] < lcs[i+1][j]) {
			sb.WriteString(fmt.Sprintf("- %s\n", origLines[i]))
			i++
		}
	}

	return sb.String(), nil
}

func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	// Normalizar terminaciones de línea Windows/Unix
	normalized := strings.ReplaceAll(s, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Split(normalized, "\n")
	// Si la última línea es vacía por un salto al final, omitir
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func computeLCS(a, b []string) [][]int {
	n, m := len(a), len(b)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}

	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else {
				if dp[i+1][j] > dp[i][j+1] {
					dp[i][j] = dp[i+1][j]
				} else {
					dp[i][j] = dp[i][j+1]
				}
			}
		}
	}
	return dp
}
