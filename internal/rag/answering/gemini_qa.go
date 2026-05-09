package answering

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/stan/Projects/studies/rag/internal/rag/retriever"
)

const (
	defaultGeminiBaseURL  = "https://generativelanguage.googleapis.com/v1beta"
	defaultGeminiLLMModel = "gemini-2.5-flash-lite"
)

// GeminiQAService gera respostas usando a Gemini Developer API.
type GeminiQAService struct {
	apiKey         string
	baseURL        string
	model          string
	client         *http.Client
	promptTemplate *PromptTemplate
	temperature    float32
	topK           int
}

type geminiGenerateRequest struct {
	Contents          []geminiContent          `json:"contents"`
	GenerationConfig  map[string]interface{}   `json:"generationConfig,omitempty"`
	SystemInstruction *geminiSystemInstruction `json:"systemInstruction,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiSystemInstruction struct {
	Parts []geminiPart `json:"parts"`
}

type geminiGenerateResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
}

// NewGeminiQAService cria um QA service para Gemini.
func NewGeminiQAService(apiKey, baseURL, model string) *GeminiQAService {
	if baseURL == "" {
		baseURL = defaultGeminiBaseURL
	}
	if model == "" {
		model = defaultGeminiLLMModel
	}

	return &GeminiQAService{
		apiKey:         apiKey,
		baseURL:        strings.TrimRight(baseURL, "/"),
		model:          strings.TrimPrefix(model, "models/"),
		client:         &http.Client{Timeout: 2 * time.Minute},
		promptTemplate: DefaultPromptTemplate(),
		temperature:    0.3,
		topK:           5,
	}
}

// WithTemperature define a temperature.
func (qs *GeminiQAService) WithTemperature(temp float32) *GeminiQAService {
	qs.temperature = temp
	return qs
}

// WithTopK define o número padrão de chunks a recuperar.
func (qs *GeminiQAService) WithTopK(k int) *GeminiQAService {
	qs.topK = k
	return qs
}

// Answer gera uma resposta para uma pergunta usando um retriever.
func (qs *GeminiQAService) Answer(ctx context.Context, question string, ret *retriever.Retriever) (*Response, error) {
	searchResults, err := ret.Retrieve(ctx, question, qs.topK)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve context: %w", err)
	}

	if len(searchResults) == 0 {
		return &Response{
			Answer: "I could not find any relevant information in the knowledge base to answer your question.",
			Model:  qs.model,
		}, nil
	}

	contextChunks := make([]string, len(searchResults))
	sources := make([]SourceInfo, len(searchResults))
	sourceRefs := make([]SourceReference, len(searchResults))

	for i, result := range searchResults {
		contextChunks[i] = result.Chunk.Content
		sources[i] = SourceInfo{
			DocumentID: result.Chunk.DocumentID,
			ChunkIndex: result.Chunk.ChunkIndex,
			Score:      result.Score,
		}
		sourceRefs[i] = SourceReference{
			DocumentID: result.Chunk.DocumentID,
			ChunkIndex: result.Chunk.ChunkIndex,
			Score:      result.Score,
			Preview:    truncateText(result.Chunk.Content, 100),
		}
	}

	prompt, err := BuildPrompt(qs.promptTemplate, question, FormatContext(contextChunks), sources)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	answer, err := qs.generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate answer: %w", err)
	}

	return &Response{
		Answer:  answer,
		Sources: sourceRefs,
		Model:   qs.model,
	}, nil
}

func (qs *GeminiQAService) generate(ctx context.Context, prompt string) (string, error) {
	payload, err := json.Marshal(geminiGenerateRequest{
		SystemInstruction: &geminiSystemInstruction{
			Parts: []geminiPart{{Text: SystemPrompt()}},
		},
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: prompt}},
			},
		},
		GenerationConfig: map[string]interface{}{
			"temperature": qs.temperature,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal Gemini generate request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		qs.endpoint("generateContent"),
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini generate request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", qs.apiKey)

	resp, err := qs.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Gemini generate API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Gemini generate API returned status %d", resp.StatusCode)
	}

	var result geminiGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Gemini generate response: %w", err)
	}
	if len(result.Candidates) == 0 {
		return "", fmt.Errorf("Gemini returned no candidates")
	}

	var answer strings.Builder
	for _, part := range result.Candidates[0].Content.Parts {
		answer.WriteString(part.Text)
	}

	text := strings.TrimSpace(answer.String())
	if text == "" {
		return "", fmt.Errorf("Gemini returned empty response")
	}

	return text, nil
}

func (qs *GeminiQAService) endpoint(method string) string {
	return fmt.Sprintf("%s/models/%s:%s", qs.baseURL, url.PathEscape(qs.model), method)
}
