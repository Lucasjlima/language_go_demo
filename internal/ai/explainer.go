package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"language_go_demo/internal/config"
)

const geminiAPIBaseURL = "https://generativelanguage.googleapis.com/v1beta/models/"

type Service struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewService(cfg config.Config) (*Service, error) {
	if strings.TrimSpace(cfg.GeminiAPIKey) == "" {
		return nil, errors.New("missing GEMINI_API_KEY in .env")
	}

	model := strings.TrimSpace(cfg.GeminiModel)
	if model == "" {
		model = "gemini-2.5-flash"
	}

	return &Service{
		apiKey:  cfg.GeminiAPIKey,
		model:   model,
		baseURL: geminiAPIBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (s *Service) Explain(req ExplanationRequest) (ExplanationResponse, error) {
	text := strings.TrimSpace(req.Text)
	if text == "" {
		return ExplanationResponse{}, ErrEmptyText
	}

	log.Printf("ai: sending request to Gemini model=%q", s.model)

	payload := geminiGenerateContentRequest{
		SystemInstruction: geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{{
					Text: userPrompt(text),
				}},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:      0.3,
			ResponseMimeType: "application/json",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return ExplanationResponse{}, fmtError("encode Gemini request", err)
	}

	respBody, err := s.generateContent(body)
	if err != nil {
		return ExplanationResponse{}, err
	}

	explanation, err := parseGeminiResponse(text, respBody)
	if err != nil {
		return ExplanationResponse{}, err
	}

	log.Printf("ai: Gemini response parsed successfully")
	return explanation, nil
}

func (s *Service) generateContent(body []byte) ([]byte, error) {
	endpoint := s.baseURL + url.PathEscape(s.model) + ":generateContent?key=" + url.QueryEscape(s.apiKey)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmtError("create Gemini request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("ai: Gemini request failed: %v", err)
		return nil, fmtError("call Gemini", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmtError("read Gemini response", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		log.Printf("ai: Gemini returned status %d", resp.StatusCode)
		return nil, fmt.Errorf("Gemini request failed: %s", extractGeminiError(respBody, resp.Status))
	}

	if len(bytes.TrimSpace(respBody)) == 0 {
		return nil, errors.New("Gemini returned an empty response")
	}

	return respBody, nil
}

func parseGeminiResponse(originalText string, body []byte) (ExplanationResponse, error) {
	var apiResp geminiGenerateContentResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return ExplanationResponse{}, fmtError("decode Gemini response", err)
	}

	if len(apiResp.Candidates) == 0 || len(apiResp.Candidates[0].Content.Parts) == 0 {
		return ExplanationResponse{}, errors.New("Gemini returned no explanation candidates")
	}

	rawJSON := cleanJSON(apiResp.Candidates[0].Content.Parts[0].Text)
	if rawJSON == "" {
		return ExplanationResponse{}, errors.New("Gemini returned an empty explanation payload")
	}

	var parsed ExplanationResponse
	if err := json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
		return ExplanationResponse{}, fmtError("convert Gemini explanation payload", err)
	}

	parsed.OriginalText = strings.TrimSpace(parsed.OriginalText)
	parsed.Explanation = strings.TrimSpace(parsed.Explanation)
	parsed.Tone = strings.TrimSpace(parsed.Tone)

	if parsed.OriginalText == "" {
		parsed.OriginalText = originalText
	}

	if parsed.Explanation == "" {
		return ExplanationResponse{}, errors.New("Gemini explanation is missing the explanation field")
	}

	if parsed.Tone == "" {
		return ExplanationResponse{}, errors.New("Gemini explanation is missing the tone field")
	}

	if len(parsed.Examples) == 0 {
		return ExplanationResponse{}, errors.New("Gemini explanation is missing examples")
	}

	filteredExamples := make([]string, 0, len(parsed.Examples))
	for _, example := range parsed.Examples {
		example = strings.TrimSpace(example)
		if example != "" {
			filteredExamples = append(filteredExamples, example)
		}
	}

	if len(filteredExamples) == 0 {
		return ExplanationResponse{}, errors.New("Gemini explanation examples were empty")
	}

	parsed.Examples = filteredExamples
	return parsed, nil
}

func cleanJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	return strings.TrimSpace(raw)
}

func extractGeminiError(body []byte, fallback string) string {
	var apiErr struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &apiErr); err == nil && strings.TrimSpace(apiErr.Error.Message) != "" {
		return apiErr.Error.Message
	}

	return fallback
}

func fmtError(action string, err error) error {
	return fmt.Errorf("%s: %w", action, err)
}

func userPrompt(text string) string {
	return "Explain this English text for a language learner in real-world context.\n\nText: " + text
}

const systemPrompt = `You are a language-learning assistant focused on helping users understand English in real context.

Explain meaning, tone, and intention clearly and naturally.
Do not default to literal translation.
Teach how the phrase is actually used by native speakers.
Keep explanations practical, concise, and easy to understand.
Provide valid JSON only with this shape:
{
  "originalText": "string",
  "explanation": "string",
  "tone": "string",
  "examples": ["string", "string", "string"]
}

Rules:
- Preserve the user's original text in originalText.
- explanation must describe meaning and likely context of use.
- tone must be a short label such as "surprised", "excited", "curious", or "neutral".
- examples must contain 2 to 4 short example sentences or similar expressions.
- Return JSON only. No markdown, no code fences, no extra keys.`

type geminiGenerateContentRequest struct {
	SystemInstruction geminiContent          `json:"system_instruction"`
	Contents          []geminiContent        `json:"contents"`
	GenerationConfig  geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature      float64 `json:"temperature"`
	ResponseMimeType string  `json:"responseMimeType"`
}

type geminiGenerateContentResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}
