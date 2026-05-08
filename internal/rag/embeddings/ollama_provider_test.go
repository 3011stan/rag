package embeddings

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

type fakeRoundTripper struct {
	status int
	body   string
	path   string
}

func (f *fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	f.path = req.URL.Path
	return &http.Response{
		StatusCode: f.status,
		Body:       io.NopCloser(bytes.NewBufferString(f.body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func TestOllamaProvider_EmbedSingle(t *testing.T) {
	transport := &fakeRoundTripper{status: http.StatusOK, body: `{"embedding":[0.1,0.2,0.3]}`}
	provider := NewOllamaProvider("http://ollama.test", "nomic-embed-text")
	provider.client.Transport = transport

	embedding, err := provider.EmbedSingle(context.Background(), "hello")
	if err != nil {
		t.Fatalf("EmbedSingle failed: %v", err)
	}

	if len(embedding) != 3 {
		t.Fatalf("expected embedding length 3, got %d", len(embedding))
	}
	if transport.path != "/api/embeddings" {
		t.Fatalf("unexpected path: %s", transport.path)
	}
}

func TestOllamaProvider_HTTPError(t *testing.T) {
	transport := &fakeRoundTripper{status: http.StatusNotFound, body: "model not found"}
	provider := NewOllamaProvider("http://ollama.test", "missing-model")
	provider.client.Transport = transport

	_, err := provider.EmbedSingle(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOllamaProvider_InvalidResponse(t *testing.T) {
	transport := &fakeRoundTripper{status: http.StatusOK, body: `{"embedding":[]}`}
	provider := NewOllamaProvider("http://ollama.test", "nomic-embed-text")
	provider.client.Transport = transport

	_, err := provider.EmbedSingle(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error")
	}
}
