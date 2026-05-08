package embeddings

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// Provider define a interface para geração de embeddings
type Provider interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	EmbedSingle(ctx context.Context, text string) ([]float32, error)
}

// OpenAIProvider é a implementação usando OpenAI
type OpenAIProvider struct {
	client    *openai.Client
	model     openai.EmbeddingModel
	batchSize int
}

// NewOpenAIProvider cria um novo provider OpenAI
func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		client:    openai.NewClient(apiKey),
		model:     openai.SmallEmbedding3,
		batchSize: 64,
	}
}

// EmbedSingle gera embedding para um texto único
func (p *OpenAIProvider) EmbedSingle(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := p.Embed(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embeddings[0], nil
}

// Embed gera embeddings para múltiplos textos em lotes
func (p *OpenAIProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	var allEmbeddings [][]float32

	// Processar em lotes
	for i := 0; i < len(texts); i += p.batchSize {
		end := i + p.batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		embeddings, err := p.embedBatch(ctx, batch)
		if err != nil {
			return nil, err
		}

		allEmbeddings = append(allEmbeddings, embeddings...)
	}

	return allEmbeddings, nil
}

// embedBatch processa um lote de textos com retry
func (p *OpenAIProvider) embedBatch(ctx context.Context, batch []string) ([][]float32, error) {
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := p.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Input: batch,
			Model: p.model,
		})

		if err == nil {
			// Extrair embeddings da resposta
			var embeddings [][]float32
			for _, item := range resp.Data {
				embeddings = append(embeddings, item.Embedding)
			}
			return embeddings, nil
		}

		// Se for a última tentativa, retornar erro
		if attempt == maxRetries-1 {
			return nil, fmt.Errorf("failed to create embeddings after %d attempts: %w", maxRetries, err)
		}

		// Aguardar antes de tentar novamente
		continue
	}

	return nil, fmt.Errorf("failed to create embeddings")
}
