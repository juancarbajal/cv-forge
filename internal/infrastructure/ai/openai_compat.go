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

type OpenAICompatibleProvider struct {
	providerName string
	baseURL      string
	apiKey       string
	model        string
	httpClient   *http.Client
}

func NewOpenAICompatibleProvider(providerName, baseURL, apiKey, model string) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{
		providerName: providerName,
		baseURL:      baseURL,
		apiKey:       apiKey,
		model:        model,
		httpClient: &http.Client{
			Timeout: 55 * time.Second,
		},
	}
}

func (p *OpenAICompatibleProvider) ProviderName() string {
	return p.providerName
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponseFormat struct {
	Type string `json:"type"`
}

type openAIChatRequest struct {
	Model          string                `json:"model"`
	Messages       []openAIChatMessage   `json:"messages"`
	Temperature    float64               `json:"temperature"`
	ResponseFormat *openAIResponseFormat `json:"response_format,omitempty"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (p *OpenAICompatibleProvider) call(ctx context.Context, systemPrompt, userMessage string, jsonMode bool) (string, error) {
	url := fmt.Sprintf("%s/chat/completions", p.baseURL)

	reqBody := openAIChatRequest{
		Model: p.model,
		Messages: []openAIChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		Temperature: 0.1,
	}
	if jsonMode {
		reqBody.ResponseFormat = &openAIResponseFormat{Type: "json_object"}
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
	if p.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error conectando a %s: %w", p.providerName, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error leyendo respuesta de %s: %w", p.providerName, err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s API error (HTTP %d): %s", p.providerName, resp.StatusCode, string(bodyBytes))
	}

	var chatResp openAIChatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", fmt.Errorf("error deserializando respuesta de %s: %w", p.providerName, err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("%s error: %s", p.providerName, chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("%s no retornó ninguna opción", p.providerName)
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (p *OpenAICompatibleProvider) DetectLanguage(ctx context.Context, text string) (domain.Language, error) {
	sys := "Detecta el idioma principal del siguiente texto. Responde ÚNICAMENTE una de estas palabras: 'es', 'en', o 'pt'."
	ans, err := p.call(ctx, sys, text, false)
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

func (p *OpenAICompatibleProvider) ExtractAndClassifyRequirements(ctx context.Context, jobPosting string, prompt string) ([]domain.Requirement, error) {
	raw, err := p.call(ctx, prompt, jobPosting, true)
	if err != nil {
		return nil, err
	}
	_, reqs, err := parseRequirements(raw)
	return reqs, err
}

func (p *OpenAICompatibleProvider) MatchCV(ctx context.Context, cvContent string, requirements []domain.Requirement, prompt string) ([]domain.MatchResult, error) {
	reqBytes, _ := json.Marshal(requirements)
	input := fmt.Sprintf("REQUISITOS:\n%s\n\nCONTENIDO CV:\n%s", string(reqBytes), cvContent)
	raw, err := p.call(ctx, prompt, input, true)
	if err != nil {
		return nil, err
	}
	return parseMatches(raw)
}

func (p *OpenAICompatibleProvider) GenerateRecommendations(ctx context.Context, cvContent string, matches []domain.MatchResult, prompt string) (domain.Recommendations, error) {
	matchBytes, _ := json.Marshal(matches)
	input := fmt.Sprintf("MATCHING PREVIO:\n%s\n\nCONTENIDO CV:\n%s", string(matchBytes), cvContent)
	raw, err := p.call(ctx, prompt, input, true)
	if err != nil {
		return domain.Recommendations{}, err
	}
	return parseRecommendations(raw)
}

func (p *OpenAICompatibleProvider) GenerateOptimizedCV(ctx context.Context, cvContent string, jobPosting string, matches []domain.MatchResult, targetLang domain.Language, prompt string) (domain.OptimizedCVResult, error) {
	matchBytes, _ := json.Marshal(matches)
	input := fmt.Sprintf("IDIOMA OBJETIVO: %s\n\nOFERTA LABORAL:\n%s\n\nMATCHING IDENTIFICADO:\n%s\n\nCV ORIGINAL:\n%s",
		string(targetLang), jobPosting, string(matchBytes), cvContent)
	raw, err := p.call(ctx, prompt, input, true)
	if err != nil {
		return domain.OptimizedCVResult{}, err
	}
	return parseOptimization(raw, targetLang)
}
