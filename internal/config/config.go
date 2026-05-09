package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	ProviderAuto   = "auto"
	ProviderGemini = "gemini"
	ProviderOllama = "ollama"
	ProviderOpenAI = "openai"
)

type Config struct {
	// Database
	DatabaseURL string

	// AI Provider
	AIProvider          string
	EmbeddingModel      string
	LLMModel            string
	EmbeddingDimensions int

	// OpenAI
	OpenAIAPIKey string

	// Gemini
	GeminiAPIKey  string
	GeminiBaseURL string

	// Ollama
	OllamaBaseURL    string
	OllamaEmbedModel string
	OllamaLLMModel   string

	// Chunking
	ChunkTokens   int
	OverlapTokens int

	// Search
	TopK int

	// Server
	Port string

	// Application
	Env        string // development, production
	AdminToken string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:         getEnvOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/rag?sslmode=disable"),
		AIProvider:          strings.ToLower(getEnvOrDefault("AI_PROVIDER", ProviderAuto)),
		EmbeddingModel:      getEnvOrDefault("EMBEDDING_MODEL", ""),
		LLMModel:            getEnvOrDefault("LLM_MODEL", ""),
		EmbeddingDimensions: getEnvAsIntOrDefault("EMBEDDING_DIMENSIONS", 768),
		OpenAIAPIKey:        getEnvOrDefault("OPENAI_API_KEY", ""),
		GeminiAPIKey:        getEnvOrDefault("GEMINI_API_KEY", ""),
		GeminiBaseURL:       getEnvOrDefault("GEMINI_BASE_URL", "https://generativelanguage.googleapis.com/v1beta"),
		OllamaBaseURL:       getEnvOrDefault("OLLAMA_BASE_URL", "http://localhost:11434"),
		OllamaEmbedModel:    getEnvOrDefault("OLLAMA_EMBED_MODEL", "nomic-embed-text"),
		OllamaLLMModel:      getEnvOrDefault("OLLAMA_LLM_MODEL", "mistral"),
		ChunkTokens:         getEnvAsIntOrDefault("CHUNK_TOKENS", 800),
		OverlapTokens:       getEnvAsIntOrDefault("OVERLAP_TOKENS", 100),
		TopK:                getEnvAsIntOrDefault("TOP_K", 5),
		Port:                normalizePort(getEnvOrDefault("PORT", ":8080")),
		Env:                 getEnvOrDefault("ENVIRONMENT", "development"),
		AdminToken:          getEnvOrDefault("ADMIN_TOKEN", ""),
	}

	cfg.applyModelDefaults()

	// Validar configurações obrigatórias
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if err := validateDatabaseURL(cfg.DatabaseURL); err != nil {
		return nil, err
	}

	if cfg.AIProvider != ProviderAuto &&
		cfg.AIProvider != ProviderGemini &&
		cfg.AIProvider != ProviderOllama &&
		cfg.AIProvider != ProviderOpenAI {
		return nil, fmt.Errorf("AI_PROVIDER must be one of: auto, gemini, ollama, openai")
	}

	if cfg.AIProvider == ProviderGemini && cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required when AI_PROVIDER=gemini")
	}
	if cfg.AIProvider == ProviderOpenAI && cfg.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required when AI_PROVIDER=openai")
	}

	return cfg, nil
}

// ResolvedAIProvider retorna o provider efetivo usado no runtime.
func (cfg *Config) ResolvedAIProvider() string {
	if cfg.AIProvider != ProviderAuto {
		return cfg.AIProvider
	}
	if cfg.GeminiAPIKey != "" {
		return ProviderGemini
	}
	if cfg.OpenAIAPIKey != "" {
		return ProviderOpenAI
	}
	return ProviderOllama
}

func (cfg *Config) applyModelDefaults() {
	if cfg.ResolvedAIProvider() == ProviderGemini {
		if cfg.EmbeddingModel == "" {
			cfg.EmbeddingModel = "gemini-embedding-001"
		}
		if cfg.LLMModel == "" {
			cfg.LLMModel = "gemini-2.5-flash-lite"
		}
		return
	}

	if cfg.EmbeddingModel == "" {
		cfg.EmbeddingModel = cfg.OllamaEmbedModel
	}
	if cfg.LLMModel == "" {
		cfg.LLMModel = cfg.OllamaLLMModel
	}
	if cfg.OllamaEmbedModel == "" {
		cfg.OllamaEmbedModel = cfg.EmbeddingModel
	}
	if cfg.OllamaLLMModel == "" {
		cfg.OllamaLLMModel = cfg.LLMModel
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func validateDatabaseURL(databaseURL string) error {
	trimmed := strings.TrimSpace(databaseURL)
	if trimmed == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if strings.Contains(trimmed, "@") && !strings.Contains(trimmed, "://") {
		return fmt.Errorf("DATABASE_URL must include a postgres:// or postgresql:// scheme; example: postgresql://postgres.<PROJECT-REF>:<PASSWORD>@aws-0-<REGION>.pooler.supabase.com:5432/postgres?sslmode=require")
	}

	if strings.Contains(trimmed, "://") {
		parsed, err := url.Parse(trimmed)
		if err != nil {
			return fmt.Errorf("DATABASE_URL is invalid: %w", err)
		}
		if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
			return fmt.Errorf("DATABASE_URL must use postgres:// or postgresql:// scheme")
		}
		if parsed.Host == "" {
			return fmt.Errorf("DATABASE_URL must include a host")
		}
	}

	return nil
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return defaultValue
	}

	return result
}

func normalizePort(port string) string {
	port = strings.TrimSpace(port)
	if port == "" {
		return ":8080"
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	if _, err := strconv.Atoi(port); err == nil {
		return ":" + port
	}
	return port
}
