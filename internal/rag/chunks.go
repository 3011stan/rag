package rag

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pgvector/pgvector-go"
)

func (store *PGVectorStore) InsertBatch(ctx context.Context, chunks []Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	tx, err := store.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO rag_chunks (id, document_id, chunk_index, content, token_count, metadata, embedding)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			content = EXCLUDED.content,
			token_count = EXCLUDED.token_count,
			metadata = EXCLUDED.metadata,
			embedding = EXCLUDED.embedding
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, chunk := range chunks {
		metadataJSON, err := marshalJSON(chunk.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal chunk metadata: %w", err)
		}

		_, err = stmt.ExecContext(
			ctx,
			chunk.ID,
			chunk.DocumentID,
			chunk.ChunkIndex,
			chunk.Content,
			chunk.TokenCount,
			metadataJSON,
			pgvector.NewVector(chunk.Embedding),
		)
		if err != nil {
			return fmt.Errorf("failed to insert chunk %s: %w", chunk.ID, err)
		}
	}

	if err := stmt.Close(); err != nil {
		return fmt.Errorf("failed to close statement: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (store *PGVectorStore) Search(ctx context.Context, embedding []float32, topK int) ([]Chunk, error) {
	return store.SearchWithFilters(ctx, embedding, topK, nil)
}

func (store *PGVectorStore) SearchWithFilters(ctx context.Context, embedding []float32, topK int, filters map[string]interface{}) ([]Chunk, error) {
	rows, err := store.DB.QueryContext(ctx, `
		SELECT
			c.id, c.document_id, c.chunk_index, c.content, c.token_count,
			c.metadata, c.embedding, c.embedding <-> $1 AS score
		FROM rag_chunks c
		ORDER BY c.embedding <-> $1
		LIMIT $2
	`, pgvector.NewVector(embedding), topK)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	return scanChunks(rows, true)
}

func (store *PGVectorStore) GetChunksByDocumentID(ctx context.Context, documentID string) ([]Chunk, error) {
	rows, err := store.DB.QueryContext(ctx, `
		SELECT id, document_id, chunk_index, content, token_count, metadata, embedding
		FROM rag_chunks
		WHERE document_id = $1
		ORDER BY chunk_index
	`, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chunks: %w", err)
	}
	defer rows.Close()

	return scanChunks(rows, false)
}

func scanChunks(rows *sql.Rows, includeScore bool) ([]Chunk, error) {
	var chunks []Chunk

	for rows.Next() {
		chunk, err := scanChunk(rows, includeScore)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, chunk)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return chunks, nil
}

func scanChunk(rows *sql.Rows, includeScore bool) (Chunk, error) {
	var chunk Chunk
	var embeddingVector pgvector.Vector
	var metadataJSON []byte

	dest := []interface{}{
		&chunk.ID,
		&chunk.DocumentID,
		&chunk.ChunkIndex,
		&chunk.Content,
		&chunk.TokenCount,
		&metadataJSON,
		&embeddingVector,
	}
	if includeScore {
		dest = append(dest, &chunk.Score)
	}

	if err := rows.Scan(dest...); err != nil {
		return Chunk{}, fmt.Errorf("failed to scan row: %w", err)
	}

	metadata, err := unmarshalJSON(metadataJSON)
	if err != nil {
		return Chunk{}, fmt.Errorf("failed to unmarshal chunk metadata: %w", err)
	}

	chunk.Embedding = embeddingVector.Slice()
	chunk.Metadata = metadata

	return chunk, nil
}
