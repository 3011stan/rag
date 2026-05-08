package rag

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	"github.com/pgvector/pgvector-go"
)

// Chunk representa um chunk de documento armazenado
type Chunk struct {
	ID         string
	DocumentID string
	ChunkIndex int
	Content    string
	TokenCount int
	Embedding  []float32
	Score      float64 // para resultados de busca
	Metadata   map[string]interface{}
}

// Document representa um documento armazenado
type Document struct {
	ID       string
	Source   string
	Title    string
	Checksum string
	Metadata map[string]interface{}
}

// VectorStore representa a interface para operações de vector database
type VectorStore interface {
	TestConnection() error
	InsertBatch(ctx context.Context, chunks []Chunk) error
	Search(ctx context.Context, embedding []float32, topK int) ([]Chunk, error)
	SearchWithFilters(ctx context.Context, embedding []float32, topK int, filters map[string]interface{}) ([]Chunk, error)
	GetChunksByDocumentID(ctx context.Context, documentID string) ([]Chunk, error)
	InsertDocument(ctx context.Context, doc Document) error
	GetDocumentByID(ctx context.Context, documentID string) (*Document, error)
	DeleteDocument(ctx context.Context, documentID string) error
}

// PGVectorStore é a implementação de VectorStore para PostgreSQL
type PGVectorStore struct {
	DB *sql.DB
}

// NewPGVectorStore cria uma nova instância de PGVectorStore
func NewPGVectorStore(dsn string) (*PGVectorStore, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	return &PGVectorStore{DB: db}, nil
}

// TestConnection verifica a conectividade com o banco de dados
func (store *PGVectorStore) TestConnection() error {
	if err := store.DB.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	return nil
}

// InsertBatch insere múltiplos chunks em lote
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
		embedding := pgvector.NewVector(chunk.Embedding)
		metadataJSON, _ := marshalJSON(chunk.Metadata)

		_, err := stmt.ExecContext(
			ctx,
			chunk.ID,
			chunk.DocumentID,
			chunk.ChunkIndex,
			chunk.Content,
			chunk.TokenCount,
			metadataJSON,
			embedding,
		)
		if err != nil {
			return fmt.Errorf("failed to insert chunk %s: %w", chunk.ID, err)
		}
	}

	if err = stmt.Close(); err != nil {
		return fmt.Errorf("failed to close statement: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Search busca chunks por similaridade de embedding
func (store *PGVectorStore) Search(ctx context.Context, embedding []float32, topK int) ([]Chunk, error) {
	return store.SearchWithFilters(ctx, embedding, topK, nil)
}

// SearchWithFilters busca chunks com filtros adicionais
func (store *PGVectorStore) SearchWithFilters(ctx context.Context, embedding []float32, topK int, filters map[string]interface{}) ([]Chunk, error) {
	vec := pgvector.NewVector(embedding)

	query := `
		SELECT 
			c.id, c.document_id, c.chunk_index, c.content, c.token_count, 
			c.metadata, c.embedding, c.embedding <-> $1 as score
		FROM rag_chunks c
		ORDER BY c.embedding <-> $1
		LIMIT $2
	`

	rows, err := store.DB.QueryContext(ctx, query, vec, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var chunks []Chunk
	for rows.Next() {
		var chunk Chunk
		var embeddingVector pgvector.Vector
		var metadataJSON []byte

		err := rows.Scan(
			&chunk.ID,
			&chunk.DocumentID,
			&chunk.ChunkIndex,
			&chunk.Content,
			&chunk.TokenCount,
			&metadataJSON,
			&embeddingVector,
			&chunk.Score,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		chunk.Embedding = embeddingVector.Slice()
		chunk.Metadata, _ = unmarshalJSON(metadataJSON)
		chunks = append(chunks, chunk)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return chunks, nil
}

// GetChunksByDocumentID retorna todos os chunks de um documento
func (store *PGVectorStore) GetChunksByDocumentID(ctx context.Context, documentID string) ([]Chunk, error) {
	query := `
		SELECT id, document_id, chunk_index, content, token_count, metadata, embedding
		FROM rag_chunks
		WHERE document_id = $1
		ORDER BY chunk_index
	`

	rows, err := store.DB.QueryContext(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chunks: %w", err)
	}
	defer rows.Close()

	var chunks []Chunk
	for rows.Next() {
		var chunk Chunk
		var embeddingVector pgvector.Vector
		var metadataJSON []byte

		err := rows.Scan(
			&chunk.ID,
			&chunk.DocumentID,
			&chunk.ChunkIndex,
			&chunk.Content,
			&chunk.TokenCount,
			&metadataJSON,
			&embeddingVector,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		chunk.Embedding = embeddingVector.Slice()
		chunk.Metadata, _ = unmarshalJSON(metadataJSON)
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// InsertDocument insere um novo documento
func (store *PGVectorStore) InsertDocument(ctx context.Context, doc Document) error {
	metadataJSON, _ := marshalJSON(doc.Metadata)

	query := `
		INSERT INTO rag_documents (id, source, title, checksum, metadata)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (checksum) DO UPDATE SET
			title = EXCLUDED.title,
			metadata = EXCLUDED.metadata
	`

	_, err := store.DB.ExecContext(ctx, query, doc.ID, doc.Source, doc.Title, doc.Checksum, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	return nil
}

// GetDocumentByID retorna um documento pelo ID
func (store *PGVectorStore) GetDocumentByID(ctx context.Context, documentID string) (*Document, error) {
	query := `
		SELECT id, source, title, checksum, metadata
		FROM rag_documents
		WHERE id = $1
	`

	var doc Document
	var metadataJSON []byte

	err := store.DB.QueryRowContext(ctx, query, documentID).Scan(
		&doc.ID,
		&doc.Source,
		&doc.Title,
		&doc.Checksum,
		&metadataJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query document: %w", err)
	}

	doc.Metadata, _ = unmarshalJSON(metadataJSON)
	return &doc, nil
}

// DeleteDocument deleta um documento e seus chunks
func (store *PGVectorStore) DeleteDocument(ctx context.Context, documentID string) error {
	query := `DELETE FROM rag_documents WHERE id = $1`

	_, err := store.DB.ExecContext(ctx, query, documentID)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// Helper functions
func marshalJSON(data map[string]interface{}) ([]byte, error) {
	if data == nil {
		return []byte("{}"), nil
	}
	// Implementar marshal JSON real
	return []byte("{}"), nil
}

func unmarshalJSON(data []byte) (map[string]interface{}, error) {
	if len(data) == 0 {
		return make(map[string]interface{}), nil
	}
	// Implementar unmarshal JSON real
	return make(map[string]interface{}), nil
}
