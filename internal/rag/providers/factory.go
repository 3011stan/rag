package providers

import (
	"fmt"

	"github.com/stan/Projects/studies/rag/internal/config"
	"github.com/stan/Projects/studies/rag/internal/rag/embeddings"
	"github.com/stan/Projects/studies/rag/internal/rag/qa"
)

// NewEmbeddingProvider cria o provider de embeddings a partir da configuracao.
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

// NewQAService cria o provider de QA a partir da configuracao.
func NewQAService(cfg *config.Config) (qa.Service, error) {
	switch cfg.ResolvedAIProvider() {
	case config.ProviderGemini:
		return qa.NewGeminiQAService(cfg.GeminiAPIKey, cfg.GeminiBaseURL, cfg.LLMModel).
			WithTemperature(0.3).
			WithTopK(cfg.TopK), nil
	case config.ProviderOllama:
		return qa.NewOllamaQAService(cfg.OllamaBaseURL, cfg.LLMModel).
			WithTemperature(0.3).
			WithTopK(cfg.TopK), nil
	case config.ProviderOpenAI:
		return qa.NewQAService(cfg.OpenAIAPIKey).
			WithTemperature(0.3).
			WithTopK(cfg.TopK), nil
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", cfg.ResolvedAIProvider())
	}
}
