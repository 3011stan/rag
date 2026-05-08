package qa

import (
	"context"
	"net/http"
	"testing"

	"github.com/stan/Projects/studies/rag/internal/rag"
	"github.com/stan/Projects/studies/rag/internal/rag/retriever"
)

func TestGeminiQAService_Answer(t *testing.T) {
	transport := &fakeRoundTripper{
		status: http.StatusOK,
		body:   `{"candidates":[{"content":{"parts":[{"text":"A resposta vem do Gemini."}]}}]}`,
	}
	vs := fakeVectorStore{
		chunks: []rag.Chunk{
			{
				ID:         "chunk-1",
				DocumentID: "doc-1",
				ChunkIndex: 0,
				Content:    "O projeto pode usar Gemini em deploy.",
				Score:      0.1,
			},
		},
	}
	ret := retriever.NewRetriever(vs, fakeEmbeddingProvider{}, 1)
	service := NewGeminiQAService("test-key", "http://gemini.test/v1beta", "gemini-2.5-flash-lite").WithTopK(1)
	service.client.Transport = transport

	resp, err := service.Answer(context.Background(), "Qual provider de deploy?", ret)
	if err != nil {
		t.Fatalf("Answer failed: %v", err)
	}
	if resp.Answer == "" {
		t.Fatal("expected answer")
	}
	if len(resp.Sources) != 1 {
		t.Fatalf("expected one source, got %d", len(resp.Sources))
	}
	if transport.path != "/v1beta/models/gemini-2.5-flash-lite:generateContent" {
		t.Fatalf("unexpected path: %s", transport.path)
	}
}

func TestGeminiQAService_HTTPError(t *testing.T) {
	transport := &fakeRoundTripper{status: http.StatusUnauthorized, body: "bad key"}
	vs := fakeVectorStore{
		chunks: []rag.Chunk{{ID: "chunk-1", DocumentID: "doc-1", Content: "context"}},
	}
	ret := retriever.NewRetriever(vs, fakeEmbeddingProvider{}, 1)
	service := NewGeminiQAService("bad-key", "http://gemini.test/v1beta", "gemini-2.5-flash-lite")
	service.client.Transport = transport

	_, err := service.Answer(context.Background(), "pergunta", ret)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGeminiQAService_InvalidResponse(t *testing.T) {
	transport := &fakeRoundTripper{status: http.StatusOK, body: `{"candidates":[]}`}
	vs := fakeVectorStore{
		chunks: []rag.Chunk{{ID: "chunk-1", DocumentID: "doc-1", Content: "context"}},
	}
	ret := retriever.NewRetriever(vs, fakeEmbeddingProvider{}, 1)
	service := NewGeminiQAService("test-key", "http://gemini.test/v1beta", "gemini-2.5-flash-lite")
	service.client.Transport = transport

	_, err := service.Answer(context.Background(), "pergunta", ret)
	if err == nil {
		t.Fatal("expected error")
	}
}
