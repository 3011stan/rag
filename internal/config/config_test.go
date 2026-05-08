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
