package ai

import (
	"fmt"

	"github.com/stan/Projects/studies/rag/internal/config"
	"github.com/stan/Projects/studies/rag/internal/rag/answering"
	"github.com/stan/Projects/studies/rag/internal/rag/embeddings"
)

func NewEmbeddingProvider(cfg *config.Config) (embeddings.Provider, error) {
	switch cfg.ResolvedAIProvider() {
	case config.ProviderGemini:
		return embeddings.NewGeminiProvider(
			cfg.GeminiAPIKey,
			cfg.GeminiBaseURL,
			cfg.EmbeddingModel,
			cfg.EmbeddingDimensions,
		), nil
	case config.ProviderOllama:
		return embeddings.NewOllamaProvider(cfg.OllamaBaseURL, cfg.EmbeddingModel), nil
	case config.ProviderOpenAI:
		return embeddings.NewOpenAIProvider(cfg.OpenAIAPIKey), nil
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", cfg.ResolvedAIProvider())
	}
}

func NewAnsweringService(cfg *config.Config) (answering.Service, error) {
	switch cfg.ResolvedAIProvider() {
	case config.ProviderGemini:
		return answering.NewGeminiQAService(cfg.GeminiAPIKey, cfg.GeminiBaseURL, cfg.LLMModel).
			WithTemperature(0.3).
			WithTopK(cfg.TopK), nil
	case config.ProviderOllama:
		return answering.NewOllamaQAService(cfg.OllamaBaseURL, cfg.LLMModel).
			WithTemperature(0.3).
			WithTopK(cfg.TopK), nil
	case config.ProviderOpenAI:
		return answering.NewQAService(cfg.OpenAIAPIKey).
			WithTemperature(0.3).
			WithTopK(cfg.TopK), nil
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", cfg.ResolvedAIProvider())
	}
}
