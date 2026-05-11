package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stan/Projects/studies/rag/internal/config"
	"github.com/stan/Projects/studies/rag/internal/rag"
)

type fakeDocumentStore struct {
	documents []rag.DocumentSummary
	deletedID string
}

func (f *fakeDocumentStore) TestConnection() error { return nil }
func (f *fakeDocumentStore) InsertBatch(ctx context.Context, chunks []rag.Chunk) error {
	return nil
}
func (f *fakeDocumentStore) Search(ctx context.Context, embedding []float32, topK int) ([]rag.Chunk, error) {
	return nil, nil
}
func (f *fakeDocumentStore) SearchWithFilters(ctx context.Context, embedding []float32, topK int, filters map[string]interface{}) ([]rag.Chunk, error) {
	return nil, nil
}
func (f *fakeDocumentStore) GetChunksByDocumentID(ctx context.Context, documentID string) ([]rag.Chunk, error) {
	return nil, nil
}
func (f *fakeDocumentStore) InsertDocument(ctx context.Context, doc rag.Document) error { return nil }
func (f *fakeDocumentStore) GetDocumentByID(ctx context.Context, documentID string) (*rag.Document, error) {
	return nil, nil
}
func (f *fakeDocumentStore) ListDocuments(ctx context.Context) ([]rag.DocumentSummary, error) {
	return f.documents, nil
}
func (f *fakeDocumentStore) DeleteDocument(ctx context.Context, documentID string) error {
	f.deletedID = documentID
	return nil
}

func TestListDocumentsAllowsTemporaryToken(t *testing.T) {
	store := &fakeDocumentStore{
		documents: []rag.DocumentSummary{
			{
				ID:         "doc-1",
				Title:      "Demo",
				ChunkCount: 3,
				CreatedAt:  time.Now(),
			},
		},
	}
	server := &APIServer{
		cfg: &config.Config{TemporaryTokenSecret: "test-temp-secret"},
		vs:  store,
	}
	token, _, err := server.createTemporaryToken("user@example.com")
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	req := httptest.NewRequest("GET", "/documents", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	server.ListDocumentsHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteDocumentRejectsTemporaryToken(t *testing.T) {
	store := &fakeDocumentStore{}
	server := &APIServer{
		cfg: &config.Config{
			AdminToken:           "admin-token",
			TemporaryTokenSecret: "test-temp-secret",
		},
		vs: store,
	}
	token, _, err := server.createTemporaryToken("user@example.com")
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/documents/doc-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	server.DeleteDocumentHandler().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
	if store.deletedID != "" {
		t.Fatalf("temporary token must not delete documents; deleted %q", store.deletedID)
	}
}

func TestDebugMetadataRejectsTemporaryToken(t *testing.T) {
	server := &APIServer{
		cfg: &config.Config{
			AdminToken:           "admin-token",
			TemporaryTokenSecret: "test-temp-secret",
		},
	}
	token, _, err := server.createTemporaryToken("user@example.com")
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	req := httptest.NewRequest("GET", "/debug/metadata", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	server.DebugMetadataHandler().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}
