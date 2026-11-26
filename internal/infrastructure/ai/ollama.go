package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jcarbajal/cv-checker/internal/domain"
)

type OllamaProvider struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewOllamaProvider(baseURL, model string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if model == "" {
		model = "llama3"
	}
	return &OllamaProvider{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 55 * time.Second,
		},
	}
}

func (o *OllamaProvider) ProviderName() string {
	return "ollama"
}

type ollamaChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatRequest struct {
	Model    string              `json:"model"`
	Messages []ollamaChatMessage `json:"messages"`
	Format   string              `json:"format,omitempty"`
	Stream   bool                `json:"stream"`
}

type ollamaChatResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Error string `json:"error,omitempty"`
}

func (o *OllamaProvider) call(ctx context.Context, systemPrompt, userMessage string, jsonMode bool) (string, error) {
	url := fmt.Sprintf("%s/api/chat", o.baseURL)

	reqBody := ollamaChatRequest{
		Model: o.model,
		Messages: []ollamaChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		Stream: false,
	}
	if jsonMode {
		reqBody.Format = "json"
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

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error conectando a Ollama local (%s): %w", o.baseURL, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error leyendo respuesta de Ollama: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama error (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var oResp ollamaChatResponse
	if err := json.Unmarshal(bodyBytes, &oResp); err != nil {
		return "", fmt.Errorf("error parseando respuesta de Ollama: %w", err)
	}

	if oResp.Error != "" {
		return "", fmt.Errorf("Ollama error: %s", oResp.Error)
	}

	return oResp.Message.Content, nil
}

func (o *OllamaProvider) DetectLanguage(ctx context.Context, text string) (domain.Language, error) {
	sys := "Detecta el idioma principal del siguiente texto. Responde ÚNICAMENTE una de estas palabras: 'es', 'en', o 'pt'."
	ans, err := o.call(ctx, sys, text, false)
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

func (o *OllamaProvider) ExtractAndClassifyRequirements(ctx context.Context, jobPosting string, prompt string) ([]domain.Requirement, error) {
	raw, err := o.call(ctx, prompt, jobPosting, true)
	if err != nil {
		return nil, err
	}
	_, reqs, err := parseRequirements(raw)
	return reqs, err
}

func (o *OllamaProvider) MatchCV(ctx context.Context, cvContent string, requirements []domain.Requirement, prompt string) ([]domain.MatchResult, error) {
	reqBytes, _ := json.Marshal(requirements)
	input := fmt.Sprintf("REQUISITOS:\n%s\n\nCONTENIDO CV:\n%s", string(reqBytes), cvContent)
	raw, err := o.call(ctx, prompt, input, true)
	if err != nil {
		return nil, err
	}
	return parseMatches(raw)
}

func (o *OllamaProvider) GenerateRecommendations(ctx context.Context, cvContent string, matches []domain.MatchResult, prompt string) (domain.Recommendations, error) {
	matchBytes, _ := json.Marshal(matches)
	input := fmt.Sprintf("MATCHING PREVIO:\n%s\n\nCONTENIDO CV:\n%s", string(matchBytes), cvContent)
	raw, err := o.call(ctx, prompt, input, true)
	if err != nil {
		return domain.Recommendations{}, err
	}
	return parseRecommendations(raw)
}

func (o *OllamaProvider) GenerateOptimizedCV(ctx context.Context, cvContent string, jobPosting string, matches []domain.MatchResult, targetLang domain.Language, prompt string) (domain.OptimizedCVResult, error) {
	matchBytes, _ := json.Marshal(matches)
	input := fmt.Sprintf("IDIOMA OBJETIVO: %s\n\nOFERTA LABORAL:\n%s\n\nMATCHING IDENTIFICADO:\n%s\n\nCV ORIGINAL:\n%s",
		string(targetLang), jobPosting, string(matchBytes), cvContent)
	raw, err := o.call(ctx, prompt, input, true)
	if err != nil {
		return domain.OptimizedCVResult{}, err
	}
	return parseOptimization(raw, targetLang)
}
