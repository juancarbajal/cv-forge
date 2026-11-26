package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jcarbajal/cv-checker/internal/domain"
)

func TestGeminiProvider_MockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Respuesta estándar Gemini con candidatos
		response := `{
			"candidates": [
				{
					"content": {
						"parts": [
							{
								"text": "{\"detected_language\": \"es\", \"requirements\": [{\"name\": \"Go\", \"categoria\": \"hard_skill\", \"prioridad\": \"obligatorio\"}]}"
							}
						]
					}
				}
			]
		}`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	gemini := NewGeminiProvider("dummy-key", "gemini-2.0-flash")
	gemini.SetBaseURL(server.URL)

	ctx := context.Background()

	// Probar extracción de requisitos
	reqs, err := gemini.ExtractAndClassifyRequirements(ctx, "Oferta para Go", "prompt")
	if err != nil {
		t.Fatalf("error extrayendo requisitos en Gemini: %v", err)
	}

	if len(reqs) != 1 || reqs[0].Name != "Go" {
		t.Errorf("resultado inesperado en Gemini: %+v", reqs)
	}
}

func TestOpenAICompatibleProvider_MockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validar Authorization header Bearer
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-secret-token" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := `{
			"choices": [
				{
					"message": {
						"content": "{\"matches\": [{\"requisito\": \"Go\", \"estado\": \"Cumple\", \"confianza\": 0.99, \"evidencia\": \"Exp Go\", \"explicacion\": \"OK\"}]}"
					}
				}
			]
		}`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	openai := NewOpenAICompatibleProvider("openai", server.URL, "test-secret-token", "gpt-4o")
	ctx := context.Background()

	matches, err := openai.MatchCV(ctx, "CV Go", []domain.Requirement{{Name: "Go"}}, "prompt")
	if err != nil {
		t.Fatalf("error en OpenAI compatible provider: %v", err)
	}

	if len(matches) != 1 || matches[0].Status != domain.MatchStatusFulfilled {
		t.Errorf("resultado inesperado en matches: %+v", matches)
	}
}

func TestOllamaProvider_MockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{
			"message": {
				"content": "{\"mejoras_recomendadas\": [\"Agregar métricas\"], \"riesgos\": [], \"oportunidades\": [], \"faltantes\": [], \"keywords_ats\": [\"Go\"]}"
			}
		}`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	ollama := NewOllamaProvider(server.URL, "llama3")
	ctx := context.Background()

	recs, err := ollama.GenerateRecommendations(ctx, "CV", []domain.MatchResult{}, "prompt")
	if err != nil {
		t.Fatalf("error en Ollama provider: %v", err)
	}

	if len(recs.Improvements) != 1 || recs.Improvements[0] != "Agregar métricas" {
		t.Errorf("resultado inesperado en recomendaciones de Ollama: %+v", recs)
	}
}

func TestProvider_ErrorResponses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	gemini := NewGeminiProvider("dummy-key", "gemini-2.0-flash")
	gemini.SetBaseURL(server.URL)

	_, err := gemini.DetectLanguage(context.Background(), "Texto")
	if err == nil {
		t.Errorf("esperaba error por HTTP 500 en Gemini")
	}

	openai := NewOpenAICompatibleProvider("openai", server.URL, "dummy", "gpt-4o")
	_, err = openai.DetectLanguage(context.Background(), "Texto")
	if err == nil {
		t.Errorf("esperaba error por HTTP 500 en OpenAI")
	}

	ollama := NewOllamaProvider(server.URL, "llama3")
	_, err = ollama.DetectLanguage(context.Background(), "Texto")
	if err == nil {
		t.Errorf("esperaba error por HTTP 500 en Ollama")
	}
}
