package answering

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/stan/Projects/studies/rag/internal/rag/retriever"
)

const (
	defaultOllamaBaseURL  = "http://localhost:11434"
	defaultOllamaLLMModel = "mistral"
)

// OllamaQAService gera respostas usando um modelo local servido pelo Ollama.
type OllamaQAService struct {
	baseURL          string
	model            string
	client           *http.Client
	promptTemplate   *PromptTemplate
	maxContextTokens int
	temperature      float32
	topK             int
}

type ollamaGenerateRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type ollamaGenerateResponse struct {
	Response string `json:"response"`
}

// NewOllamaQAService cria um QA service para o Ollama.
func NewOllamaQAService(baseURL, model string) *OllamaQAService {
	if baseURL == "" {
		baseURL = defaultOllamaBaseURL
	}
	if model == "" {
		model = defaultOllamaLLMModel
	}

	return &OllamaQAService{
		baseURL:          strings.TrimRight(baseURL, "/"),
		model:            model,
		client:           &http.Client{Timeout: 5 * time.Minute},
		promptTemplate:   DefaultPromptTemplate(),
		maxContextTokens: 2000,
		temperature:      0.3,
		topK:             5,
	}
}

// WithTemperature define a temperature.
func (qs *OllamaQAService) WithTemperature(temp float32) *OllamaQAService {
	qs.temperature = temp
	return qs
}

// WithTopK define o número padrão de chunks a recuperar.
func (qs *OllamaQAService) WithTopK(k int) *OllamaQAService {
	qs.topK = k
	return qs
}

// Answer gera uma resposta para uma pergunta usando um retriever.
func (qs *OllamaQAService) Answer(ctx context.Context, question string, ret *retriever.Retriever) (*Response, error) {
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

	answer, err := qs.generate(ctx, SystemPrompt()+"\n\n"+prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate answer: %w", err)
	}

	return &Response{
		Answer:  answer,
		Sources: sourceRefs,
		Model:   qs.model,
	}, nil
}

func (qs *OllamaQAService) generate(ctx context.Context, prompt string) (string, error) {
	payload, err := json.Marshal(ollamaGenerateRequest{
		Model:  qs.model,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": qs.temperature,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal Ollama generate request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		qs.baseURL+"/api/generate",
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create Ollama generate request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := qs.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama generate API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Ollama generate API returned status %d", resp.StatusCode)
	}

	var result ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Ollama generate response: %w", err)
	}

	answer := strings.TrimSpace(result.Response)
	if answer == "" {
		return "", fmt.Errorf("Ollama returned empty response")
	}

	return answer, nil
}
