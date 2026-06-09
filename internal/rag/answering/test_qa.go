package answering

import (
	"context"
	"fmt"

	"github.com/stan/Projects/studies/rag/internal/rag/retriever"
)

type TestQAService struct {
	topK int
}

func NewTestQAService() *TestQAService {
	return &TestQAService{topK: 5}
}

func (qs *TestQAService) WithTopK(k int) *TestQAService {
	qs.topK = k
	return qs
}

func (qs *TestQAService) Answer(ctx context.Context, question string, ret *retriever.Retriever) (*Response, error) {
	searchResults, err := ret.Retrieve(ctx, question, qs.topK)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve context: %w", err)
	}
	return qs.AnswerFromResults(ctx, question, searchResults)
}

func (qs *TestQAService) AnswerFromResults(ctx context.Context, question string, searchResults []retriever.Result) (*Response, error) {
	sourceRefs := make([]SourceReference, len(searchResults))
	for i, result := range searchResults {
		sourceRefs[i] = SourceReference{
			DocumentID: result.Chunk.DocumentID,
			ChunkIndex: result.Chunk.ChunkIndex,
			Score:      result.Score,
			Preview:    truncateText(result.Chunk.Content, 100),
		}
	}

	return &Response{
		Answer:  fmt.Sprintf("Deterministic eval answer generated from %d retrieved chunks.", len(searchResults)),
		Sources: sourceRefs,
		Model:   "deterministic-test-answerer",
	}, nil
}
