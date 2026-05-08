package chunker

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/pkoukk/tiktoken-go"
	"github.com/stan/Projects/studies/rag/internal/rag"
)

// ChunkerConfig contém configurações para chunking
type ChunkerConfig struct {
	ChunkTokens   int
	OverlapTokens int
}

// Chunker responsável por quebrar texto em chunks
type Chunker struct {
	config    ChunkerConfig
	tokenizer *tiktoken.Tiktoken
}

// NewChunker cria uma nova instância do Chunker
func NewChunker(config ChunkerConfig) (*Chunker, error) {
	tokenizer, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("failed to load tokenizer: %w", err)
	}

	return &Chunker{
		config:    config,
		tokenizer: tokenizer,
	}, nil
}

// ChunkText quebra um texto em chunks com overlap
func (c *Chunker) ChunkText(documentID string, text string, metadata map[string]interface{}) ([]rag.Chunk, error) {
	// Normalizar texto
	text = normalizeText(text)
	if text == "" {
		return []rag.Chunk{}, nil
	}

	// Tokenizar todo o texto
	tokens := c.tokenizer.Encode(text, nil, nil)

	var chunks []rag.Chunk
	var chunkIndex int

	// Quebrar em chunks com overlap
	for i := 0; i < len(tokens); i += (c.config.ChunkTokens - c.config.OverlapTokens) {
		end := i + c.config.ChunkTokens
		if end > len(tokens) {
			end = len(tokens)
		}

		chunkTokens := tokens[i:end]
		if len(chunkTokens) == 0 {
			continue
		}

		// Decodificar tokens para texto
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

		// Se chegou ao fim, não continuar
		if end == len(tokens) {
			break
		}
	}

	return chunks, nil
}

// Helper functions

func normalizeText(text string) string {
	// Remover caracteres de controle
	text = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, text)

	// Remover espaços extras
	text = strings.Join(strings.Fields(text), " ")

	return strings.TrimSpace(text)
}
