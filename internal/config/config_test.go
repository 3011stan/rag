package config

import "testing"

func TestResolvedAIProvider_AutoWithoutOpenAIKeyUsesOllama(t *testing.T) {
	cfg := &Config{AIProvider: ProviderAuto}

	if got := cfg.ResolvedAIProvider(); got != ProviderOllama {
		t.Fatalf("expected %s, got %s", ProviderOllama, got)
	}
}

func TestResolvedAIProvider_AutoWithOpenAIKeyUsesOpenAI(t *testing.T) {
	cfg := &Config{AIProvider: ProviderAuto, OpenAIAPIKey: "sk-test"}

	if got := cfg.ResolvedAIProvider(); got != ProviderOpenAI {
		t.Fatalf("expected %s, got %s", ProviderOpenAI, got)
	}
}

func TestResolvedAIProvider_AutoWithGeminiKeyUsesGemini(t *testing.T) {
	cfg := &Config{AIProvider: ProviderAuto, GeminiAPIKey: "gemini-test"}

	if got := cfg.ResolvedAIProvider(); got != ProviderGemini {
		t.Fatalf("expected %s, got %s", ProviderGemini, got)
	}
}

func TestResolvedAIProvider_ExplicitProviderWins(t *testing.T) {
	cfg := &Config{AIProvider: ProviderOllama, OpenAIAPIKey: "sk-test"}

	if got := cfg.ResolvedAIProvider(); got != ProviderOllama {
		t.Fatalf("expected %s, got %s", ProviderOllama, got)
	}
}

func TestLoadRejectsInvalidProvider(t *testing.T) {
	t.Setenv("AI_PROVIDER", "invalid")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid provider error")
	}
}

func TestLoadRequiresGeminiKeyForExplicitGemini(t *testing.T) {
	t.Setenv("AI_PROVIDER", ProviderGemini)
	t.Setenv("GEMINI_API_KEY", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected missing Gemini key error")
	}
}

func TestLoadRequiresOpenAIKeyForExplicitOpenAI(t *testing.T) {
	t.Setenv("AI_PROVIDER", ProviderOpenAI)
	t.Setenv("OPENAI_API_KEY", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected missing OpenAI key error")
	}
}

func TestLoadNormalizesNumericPort(t *testing.T) {
	t.Setenv("PORT", "10000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Port != ":10000" {
		t.Fatalf("expected :10000, got %s", cfg.Port)
	}
}

func TestLoadRejectsDatabaseURLWithoutScheme(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres.project-ref:password@aws-0-us-west-2.pooler.supabase.com:5432/postgres")

	if _, err := Load(); err == nil {
		t.Fatal("expected DATABASE_URL format error")
	}
}

func TestLoadAcceptsPostgresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://postgres.project-ref:password@aws-0-us-west-2.pooler.supabase.com:5432/postgres?sslmode=require")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.DatabaseURL == "" {
		t.Fatal("expected database URL")
	}
}

func TestLoadDisablesPublicUploadByDefaultInProduction(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.PublicUploadEnabled {
		t.Fatal("expected public upload disabled in production")
	}
}

func TestLoadAllowsPublicUploadOverride(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ENABLE_PUBLIC_UPLOAD", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !cfg.PublicUploadEnabled {
		t.Fatal("expected public upload override")
	}
}

func TestLoadParsesCORSAllowedOrigins(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000, https://rag-lab.vercel.app ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Fatalf("expected two origins, got %d", len(cfg.CORSAllowedOrigins))
	}
	if cfg.CORSAllowedOrigins[1] != "https://rag-lab.vercel.app" {
		t.Fatalf("unexpected origin: %q", cfg.CORSAllowedOrigins[1])
	}
}
