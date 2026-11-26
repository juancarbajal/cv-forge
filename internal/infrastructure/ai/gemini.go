package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jcarbajal/cv-checker/internal/domain"
)

type GeminiProvider struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewGeminiProvider(apiKey, model string) *GeminiProvider {
	if model == "" {
		model = "gemini-2.0-flash"
	}
	return &GeminiProvider{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 55 * time.Second, // Respetar RNF-008 (< 60s)
		},
	}
}

func (g *GeminiProvider) SetBaseURL(baseURL string) {
	g.baseURL = baseURL
}

func (g *GeminiProvider) ProviderName() string {
	return "gemini"
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiGenConfig struct {
	Temperature      float64 `json:"temperature"`
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
}

type geminiRequest struct {
	Contents         []geminiContent `json:"contents"`
	GenerationConfig geminiGenConfig `json:"generationConfig"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

func (g *GeminiProvider) callGemini(ctx context.Context, systemPrompt, userMessage string, jsonMode bool) (string, error) {
	baseURL := g.baseURL
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", baseURL, g.model, g.apiKey)

	promptText := fmt.Sprintf("%s\n\n=== ENTRADA ===\n%s", systemPrompt, userMessage)
	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: promptText}},
			},
		},
		GenerationConfig: geminiGenConfig{
			Temperature: 0.1,
		},
	}
	if jsonMode {
		reqBody.GenerationConfig.ResponseMimeType = "application/json"
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error llamando a Gemini API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error leyendo respuesta de Gemini: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini API error (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var gResp geminiResponse
	if err := json.Unmarshal(bodyBytes, &gResp); err != nil {
		return "", fmt.Errorf("error deserializando respuesta de Gemini: %w", err)
	}

	if gResp.Error != nil {
		return "", fmt.Errorf("Gemini error: %s (code %d)", gResp.Error.Message, gResp.Error.Code)
	}

	if len(gResp.Candidates) == 0 || len(gResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("Gemini no retornó contenido")
	}

	return gResp.Candidates[0].Content.Parts[0].Text, nil
}

func (g *GeminiProvider) DetectLanguage(ctx context.Context, text string) (domain.Language, error) {
	sys := "Detecta el idioma principal del siguiente texto. Responde ÚNICAMENTE una de estas palabras: 'es', 'en', o 'pt'."
	ans, err := g.callGemini(ctx, sys, text, false)
	if err != nil {
		return domain.LangSpanish, err
	}
	ansClean := CleanJSONResponse(ans)
	switch ansClean {
	case "en":
		return domain.LangEnglish, nil
	case "pt":
		return domain.LangPortuguese, nil
	default:
		return domain.LangSpanish, nil
	}
}

func (g *GeminiProvider) ExtractAndClassifyRequirements(ctx context.Context, jobPosting string, prompt string) ([]domain.Requirement, error) {
	raw, err := g.callGemini(ctx, prompt, jobPosting, true)
	if err != nil {
		return nil, err
	}
	_, reqs, err := parseRequirements(raw)
	return reqs, err
}

func (g *GeminiProvider) MatchCV(ctx context.Context, cvContent string, requirements []domain.Requirement, prompt string) ([]domain.MatchResult, error) {
	reqBytes, _ := json.Marshal(requirements)
	input := fmt.Sprintf("REQUISITOS:\n%s\n\nCONTENIDO CV:\n%s", string(reqBytes), cvContent)
	raw, err := g.callGemini(ctx, prompt, input, true)
	if err != nil {
		return nil, err
	}
	return parseMatches(raw)
}

func (g *GeminiProvider) GenerateRecommendations(ctx context.Context, cvContent string, matches []domain.MatchResult, prompt string) (domain.Recommendations, error) {
	matchBytes, _ := json.Marshal(matches)
	input := fmt.Sprintf("MATCHING PREVIO:\n%s\n\nCONTENIDO CV:\n%s", string(matchBytes), cvContent)
	raw, err := g.callGemini(ctx, prompt, input, true)
	if err != nil {
		return domain.Recommendations{}, err
	}
	return parseRecommendations(raw)
}

func (g *GeminiProvider) GenerateOptimizedCV(ctx context.Context, cvContent string, jobPosting string, matches []domain.MatchResult, targetLang domain.Language, prompt string) (domain.OptimizedCVResult, error) {
	matchBytes, _ := json.Marshal(matches)
	input := fmt.Sprintf("IDIOMA OBJETIVO: %s\n\nOFERTA LABORAL:\n%s\n\nMATCHING IDENTIFICADO:\n%s\n\nCV ORIGINAL:\n%s",
		string(targetLang), jobPosting, string(matchBytes), cvContent)
	raw, err := g.callGemini(ctx, prompt, input, true)
	if err != nil {
		return domain.OptimizedCVResult{}, err
	}
	return parseOptimization(raw, targetLang)
}
