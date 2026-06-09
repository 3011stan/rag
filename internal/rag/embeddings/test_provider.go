package embeddings

import (
	"context"
	"hash/fnv"
	"math"
	"strings"
	"unicode"
)

const defaultTestEmbeddingDimensions = 256

type TestProvider struct {
	dimensions int
}

func NewTestProvider(dimensions int) *TestProvider {
	if dimensions <= 0 {
		dimensions = defaultTestEmbeddingDimensions
	}
	return &TestProvider{dimensions: dimensions}
}

func (p *TestProvider) EmbedSingle(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := p.Embed(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return embeddings[0], nil
}

func (p *TestProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			embeddings[i] = p.embedText(text)
		}
	}
	return embeddings, nil
}

func (p *TestProvider) embedText(text string) []float32 {
	vector := make([]float32, p.dimensions)
	for _, token := range tokenize(text) {
		index := hashToken(token) % uint32(p.dimensions)
		vector[index] += 1
	}
	normalize(vector)
	return vector
}

func tokenize(text string) []string {
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	tokens := make([]string, 0, len(fields))
	for _, field := range fields {
		if field != "" {
			tokens = append(tokens, field)
		}
	}
	return tokens
}

func hashToken(token string) uint32 {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(token))
	return hasher.Sum32()
}

func normalize(vector []float32) {
	var sum float64
	for _, value := range vector {
		sum += float64(value * value)
	}
	if sum == 0 {
		return
	}
	norm := float32(math.Sqrt(sum))
	for i := range vector {
		vector[i] /= norm
	}
}
