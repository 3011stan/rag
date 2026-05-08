package qa

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stan/Projects/studies/rag/internal/rag"
	"github.com/stan/Projects/studies/rag/internal/rag/retriever"
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

type fakeEmbeddingProvider struct{}

func (fakeEmbeddingProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	return [][]float32{{0.1, 0.2, 0.3}}, nil
}

func (fakeEmbeddingProvider) EmbedSingle(ctx context.Context, text string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}

type fakeVectorStore struct {
	chunks []rag.Chunk
}

func (f fakeVectorStore) TestConnection() error { return nil }
func (f fakeVectorStore) InsertBatch(ctx context.Context, chunks []rag.Chunk) error {
	return nil
}
func (f fakeVectorStore) Search(ctx context.Context, embedding []float32, topK int) ([]rag.Chunk, error) {
	return f.chunks, nil
}
func (f fakeVectorStore) SearchWithFilters(ctx context.Context, embedding []float32, topK int, filters map[string]interface{}) ([]rag.Chunk, error) {
	return f.chunks, nil
}
func (f fakeVectorStore) GetChunksByDocumentID(ctx context.Context, documentID string) ([]rag.Chunk, error) {
	return f.chunks, nil
}
func (f fakeVectorStore) InsertDocument(ctx context.Context, doc rag.Document) error { return nil }
func (f fakeVectorStore) GetDocumentByID(ctx context.Context, documentID string) (*rag.Document, error) {
	return nil, nil
}
func (f fakeVectorStore) DeleteDocument(ctx context.Context, documentID string) error { return nil }

func TestOllamaQAService_Answer(t *testing.T) {
	transport := &fakeRoundTripper{
		status: http.StatusOK,
		body:   `{"response":"A resposta vem do contexto."}`,
	}
	vs := fakeVectorStore{
		chunks: []rag.Chunk{
			{
				ID:         "chunk-1",
				DocumentID: "doc-1",
				ChunkIndex: 0,
				Content:    "O projeto usa Ollama local.",
				Score:      0.1,
			},
		},
	}
	ret := retriever.NewRetriever(vs, fakeEmbeddingProvider{}, 1)
	service := NewOllamaQAService("http://ollama.test", "mistral").WithTopK(1)
	service.client.Transport = transport

	resp, err := service.Answer(context.Background(), "Qual modelo local?", ret)
	if err != nil {
		t.Fatalf("Answer failed: %v", err)
	}
	if resp.Answer == "" {
		t.Fatal("expected answer")
	}
	if len(resp.Sources) != 1 {
		t.Fatalf("expected one source, got %d", len(resp.Sources))
	}
	if transport.path != "/api/generate" {
		t.Fatalf("unexpected path: %s", transport.path)
	}
}

func TestOllamaQAService_HTTPError(t *testing.T) {
	transport := &fakeRoundTripper{status: http.StatusNotFound, body: "model not found"}
	vs := fakeVectorStore{
		chunks: []rag.Chunk{{ID: "chunk-1", DocumentID: "doc-1", Content: "context"}},
	}
	ret := retriever.NewRetriever(vs, fakeEmbeddingProvider{}, 1)
	service := NewOllamaQAService("http://ollama.test", "missing-model")
	service.client.Transport = transport

	_, err := service.Answer(context.Background(), "pergunta", ret)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOllamaQAService_InvalidResponse(t *testing.T) {
	transport := &fakeRoundTripper{status: http.StatusOK, body: `{"response":""}`}
	vs := fakeVectorStore{
		chunks: []rag.Chunk{{ID: "chunk-1", DocumentID: "doc-1", Content: "context"}},
	}
	ret := retriever.NewRetriever(vs, fakeEmbeddingProvider{}, 1)
	service := NewOllamaQAService("http://ollama.test", "mistral")
	service.client.Transport = transport

	_, err := service.Answer(context.Background(), "pergunta", ret)
	if err == nil {
		t.Fatal("expected error")
	}
}
