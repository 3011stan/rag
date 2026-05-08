package embeddings

import (
	"context"
	"net/http"
	"testing"
)

func TestGeminiProvider_EmbedSingle(t *testing.T) {
	transport := &fakeRoundTripper{
		status: http.StatusOK,
		body:   `{"embedding":{"values":[0.1,0.2,0.3]}}`,
	}
	provider := NewGeminiProvider("test-key", "http://gemini.test/v1beta", "gemini-embedding-001", 768)
	provider.client.Transport = transport

	embedding, err := provider.EmbedSingle(context.Background(), "hello")
	if err != nil {
		t.Fatalf("EmbedSingle failed: %v", err)
	}

	if len(embedding) != 3 {
		t.Fatalf("expected embedding length 3, got %d", len(embedding))
	}
	if transport.path != "/v1beta/models/gemini-embedding-001:embedContent" {
		t.Fatalf("unexpected path: %s", transport.path)
	}
}

func TestGeminiProvider_HTTPError(t *testing.T) {
	transport := &fakeRoundTripper{status: http.StatusUnauthorized, body: "bad key"}
	provider := NewGeminiProvider("bad-key", "http://gemini.test/v1beta", "gemini-embedding-001", 768)
	provider.client.Transport = transport

	_, err := provider.EmbedSingle(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGeminiProvider_InvalidResponse(t *testing.T) {
	transport := &fakeRoundTripper{status: http.StatusOK, body: `{"embedding":{"values":[]}}`}
	provider := NewGeminiProvider("test-key", "http://gemini.test/v1beta", "gemini-embedding-001", 768)
	provider.client.Transport = transport

	_, err := provider.EmbedSingle(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error")
	}
}
