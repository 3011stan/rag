package pipeline

import (
	"context"
	"fmt"

	"github.com/stan/Projects/studies/rag/internal/logging"
	"github.com/stan/Projects/studies/rag/internal/rag"
	"github.com/stan/Projects/studies/rag/internal/rag/answering"
	"github.com/stan/Projects/studies/rag/internal/rag/chunker"
	"github.com/stan/Projects/studies/rag/internal/rag/embeddings"
	"github.com/stan/Projects/studies/rag/internal/rag/loader"
	"github.com/stan/Projects/studies/rag/internal/rag/retriever"
)

type Pipeline struct {
	store     rag.VectorStore
	embedder  embeddings.Provider
	chunker   *chunker.Chunker
	pdfLoader *loader.PDFLoader
	retriever *retriever.Retriever
	answerer  answering.Service
}

type IngestResult struct {
	DocumentID string
	ChunkCount int
}

type AnswerResult struct {
	Answer  string
	Sources []Source
}

type Source struct {
	DocumentID string
	ChunkIndex int
	Score      float64
	Preview    string
}

func New(
	store rag.VectorStore,
	embedder embeddings.Provider,
	chunker *chunker.Chunker,
	pdfLoader *loader.PDFLoader,
	retriever *retriever.Retriever,
	answerer answering.Service,
) *Pipeline {
	return &Pipeline{
		store:     store,
		embedder:  embedder,
		chunker:   chunker,
		pdfLoader: pdfLoader,
		retriever: retriever,
		answerer:  answerer,
	}
}

func (p *Pipeline) IngestPDF(ctx context.Context, fileData []byte, fileName string) (*IngestResult, error) {
	logger := logging.FromContext(ctx)

	doc, text, err := p.pdfLoader.LoadPDF(fileData, fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PDF: %w", err)
	}

	logger.Info().
		Int("char_count", len(text)).
		Str("filename", fileName).
		Msg("PDF text extracted")

	if err := p.store.InsertDocument(ctx, *doc); err != nil {
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}

	chunks, err := p.chunker.ChunkText(doc.ID, text, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk text: %w", err)
	}
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks generated from text")
	}

	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}

	embeddings, err := p.embedder.Embed(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}
	if len(embeddings) != len(chunks) {
		return nil, fmt.Errorf("embedding count mismatch")
	}

	for i := range chunks {
		chunks[i].Embedding = embeddings[i]
	}

	if err := p.store.InsertBatch(ctx, chunks); err != nil {
		return nil, fmt.Errorf("failed to insert chunks: %w", err)
	}

	logger.Info().
		Str("document_id", doc.ID).
		Int("chunk_count", len(chunks)).
		Msg("document ingestion completed")

	return &IngestResult{
		DocumentID: doc.ID,
		ChunkCount: len(chunks),
	}, nil
}

func (p *Pipeline) Ask(ctx context.Context, question string, topK int) (*AnswerResult, error) {
	logger := logging.FromContext(ctx)

	searchResults, err := p.retriever.Retrieve(ctx, question, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve context: %w", err)
	}
	if len(searchResults) == 0 {
		logger.Warn().Msg("no relevant chunks found for question")
		return &AnswerResult{
			Answer:  "I could not find any relevant information in the knowledge base to answer your question.",
			Sources: []Source{},
		}, nil
	}

	sources := make([]Source, len(searchResults))
	for i, result := range searchResults {
		sources[i] = Source{
			DocumentID: result.Chunk.DocumentID,
			ChunkIndex: result.Chunk.ChunkIndex,
			Score:      result.Score,
			Preview:    truncate(result.Chunk.Content, 100),
		}
	}

	answer, err := p.answerer.Answer(ctx, question, p.retriever)
	if err != nil {
		return nil, fmt.Errorf("failed to generate answer: %w", err)
	}

	logger.Info().
		Int("source_count", len(sources)).
		Msg("answer generated")

	return &AnswerResult{
		Answer:  answer.Answer,
		Sources: sources,
	}, nil
}

func truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
