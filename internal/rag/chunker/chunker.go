package chunker

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/pkoukk/tiktoken-go"
	"github.com/stan/Projects/studies/rag/internal/rag"
)

type ChunkerConfig struct {
	ChunkTokens   int
	OverlapTokens int
}

type Chunker struct {
	config    ChunkerConfig
	tokenizer *tiktoken.Tiktoken
}

func NewChunker(config ChunkerConfig) (*Chunker, error) {
	if config.ChunkTokens <= 0 {
		return nil, fmt.Errorf("chunk tokens must be positive")
	}
	if config.OverlapTokens < 0 || config.OverlapTokens >= config.ChunkTokens {
		return nil, fmt.Errorf("overlap tokens must be non-negative and smaller than chunk tokens")
	}

	tokenizer, _ := tiktoken.GetEncoding("cl100k_base")
	return &Chunker{
		config:    config,
		tokenizer: tokenizer,
	}, nil
}

func (c *Chunker) ChunkText(documentID string, text string, metadata map[string]interface{}) ([]rag.Chunk, error) {
	text = normalizeText(text)
	if text == "" {
		return []rag.Chunk{}, nil
	}
	if c.tokenizer == nil {
		return c.chunkWords(documentID, text, metadata), nil
	}

	tokens := c.tokenizer.Encode(text, nil, nil)

	var chunks []rag.Chunk
	var chunkIndex int

	for i := 0; i < len(tokens); i += (c.config.ChunkTokens - c.config.OverlapTokens) {
		end := i + c.config.ChunkTokens
		if end > len(tokens) {
			end = len(tokens)
		}

		chunkTokens := tokens[i:end]
		if len(chunkTokens) == 0 {
			continue
		}

		chunkText := c.tokenizer.Decode(chunkTokens)
		chunkText = strings.TrimSpace(chunkText)

		if chunkText != "" {
			chunk := rag.Chunk{
				ID:         uuid.New().String(),
				DocumentID: documentID,
				ChunkIndex: chunkIndex,
				Content:    chunkText,
				TokenCount: len(chunkTokens),
				Metadata:   metadata,
			}
			chunks = append(chunks, chunk)
			chunkIndex++
		}

		if end == len(tokens) {
			break
		}
	}

	return chunks, nil
}

func (c *Chunker) chunkWords(documentID, text string, metadata map[string]interface{}) []rag.Chunk {
	words := strings.Fields(text)
	step := c.config.ChunkTokens - c.config.OverlapTokens
	chunks := make([]rag.Chunk, 0, (len(words)/step)+1)

	for start, index := 0, 0; start < len(words); start, index = start+step, index+1 {
		end := start + c.config.ChunkTokens
		if end > len(words) {
			end = len(words)
		}

		chunks = append(chunks, rag.Chunk{
			ID:         uuid.New().String(),
			DocumentID: documentID,
			ChunkIndex: index,
			Content:    strings.Join(words[start:end], " "),
			TokenCount: end - start,
			Metadata:   metadata,
		})
		if end == len(words) {
			break
		}
	}

	return chunks
}

func normalizeText(text string) string {
	text = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, text)

	text = strings.Join(strings.Fields(text), " ")

	return strings.TrimSpace(text)
}
