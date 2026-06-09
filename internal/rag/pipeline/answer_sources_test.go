package pipeline

import (
	"context"
	"testing"

	"github.com/stan/Projects/studies/rag/internal/rag"
	"github.com/stan/Projects/studies/rag/internal/rag/answering"
	"github.com/stan/Projects/studies/rag/internal/rag/retriever"
)

type fakeAnswerStore struct {
	chunks []rag.Chunk
}

func (f *fakeAnswerStore) TestConnection() error { return nil }
func (f *fakeAnswerStore) InsertBatch(context.Context, []rag.Chunk) error {
	return nil
}
func (f *fakeAnswerStore) Search(context.Context, []float32, int) ([]rag.Chunk, error) {
	return f.chunks, nil
}
func (f *fakeAnswerStore) SearchWithFilters(context.Context, []float32, int, map[string]interface{}) ([]rag.Chunk, error) {
	return f.chunks, nil
}
func (f *fakeAnswerStore) GetChunksByDocumentID(context.Context, string) ([]rag.Chunk, error) {
	return nil, nil
}
func (f *fakeAnswerStore) InsertDocument(context.Context, rag.Document) error { return nil }
func (f *fakeAnswerStore) GetDocumentByID(context.Context, string) (*rag.Document, error) {
	return nil, nil
}
func (f *fakeAnswerStore) ListDocuments(context.Context) ([]rag.DocumentSummary, error) {
	return nil, nil
}
func (f *fakeAnswerStore) DeleteDocument(context.Context, string) error { return nil }

type fakeAnswerEmbedder struct{}

func (fakeAnswerEmbedder) Embed(context.Context, []string) ([][]float32, error) {
	return [][]float32{{0.1, 0.2, 0.3}}, nil
}
func (fakeAnswerEmbedder) EmbedSingle(context.Context, string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}

type fakeAnswerer struct{}

func (fakeAnswerer) Answer(context.Context, string, *retriever.Retriever) (*answering.Response, error) {
	return &answering.Response{Answer: "answer"}, nil
}

func TestAskIncludesChunkMetadataInSources(t *testing.T) {
	store := &fakeAnswerStore{
		chunks: []rag.Chunk{
			{
				DocumentID: "doc-1",
				ChunkIndex: 2,
				Content:    "retrieved content",
				Score:      0.42,
				Metadata: map[string]interface{}{
					"layer":    "foundations",
					"checksum": "hidden-at-api-boundary",
				},
			},
		},
	}
	ret := retriever.NewRetriever(store, fakeAnswerEmbedder{}, 5)
	p := &Pipeline{
		retriever: ret,
		answerer:  fakeAnswerer{},
	}

	result, err := p.Ask(context.Background(), "question", 1)
	if err != nil {
		t.Fatalf("expected ask to succeed: %v", err)
	}
	if len(result.Sources) != 1 {
		t.Fatalf("expected one source, got %d", len(result.Sources))
	}
	if result.Sources[0].Metadata["layer"] != "foundations" {
		t.Fatalf("expected source metadata to be propagated, got %v", result.Sources[0].Metadata)
	}
}
