package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultGeminiBaseURL    = "https://generativelanguage.googleapis.com/v1beta"
	defaultGeminiEmbedModel = "gemini-embedding-001"
	defaultGeminiDimensions = 768
)

// GeminiProvider implementa embeddings usando a Gemini Developer API.
type GeminiProvider struct {
	apiKey     string
	baseURL    string
	model      string
	dimensions int
	client     *http.Client
}

type geminiEmbeddingRequest struct {
	Content              geminiContent `json:"content"`
	OutputDimensionality int           `json:"output_dimensionality,omitempty"`
}

type geminiEmbeddingResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

// NewGeminiProvider cria um provider de embeddings via Gemini.
func NewGeminiProvider(apiKey, baseURL, model string, dimensions int) *GeminiProvider {
	if baseURL == "" {
		baseURL = defaultGeminiBaseURL
	}
	if model == "" {
		model = defaultGeminiEmbedModel
	}
	if dimensions <= 0 {
		dimensions = defaultGeminiDimensions
	}

	return &GeminiProvider{
		apiKey:     apiKey,
		baseURL:    strings.TrimRight(baseURL, "/"),
		model:      strings.TrimPrefix(model, "models/"),
		dimensions: dimensions,
		client:     &http.Client{Timeout: 2 * time.Minute},
	}
}

// EmbedSingle gera embedding para um texto único.
func (p *GeminiProvider) EmbedSingle(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := p.Embed(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return embeddings[0], nil
}

// Embed gera embeddings para múltiplos textos.
func (p *GeminiProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	embeddings := make([][]float32, 0, len(texts))
	for _, text := range texts {
		embedding, err := p.embedOne(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings = append(embeddings, embedding)
	}

	return embeddings, nil
}

func (p *GeminiProvider) embedOne(ctx context.Context, text string) ([]float32, error) {
	payload, err := json.Marshal(geminiEmbeddingRequest{
		Content: geminiContent{
			Parts: []geminiPart{{Text: text}},
		},
		OutputDimensionality: p.dimensions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Gemini embedding request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.endpoint("embedContent"),
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini embedding request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Gemini embeddings API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Gemini embeddings API returned status %d", resp.StatusCode)
	}

	var result geminiEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Gemini embedding response: %w", err)
	}
	if len(result.Embedding.Values) == 0 {
		return nil, fmt.Errorf("Gemini returned empty embedding")
	}

	return result.Embedding.Values, nil
}

func (p *GeminiProvider) endpoint(method string) string {
	return fmt.Sprintf("%s/models/%s:%s", p.baseURL, url.PathEscape(p.model), method)
}
