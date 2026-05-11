package rag

import (
	"context"
	"database/sql"
	"fmt"
)

func (store *PGVectorStore) InsertDocument(ctx context.Context, doc Document) error {
	metadataJSON, err := marshalJSON(doc.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal document metadata: %w", err)
	}

	_, err = store.DB.ExecContext(ctx, `
		INSERT INTO rag_documents (id, source, title, checksum, metadata)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (checksum) DO UPDATE SET
			title = EXCLUDED.title,
			metadata = EXCLUDED.metadata
	`, doc.ID, doc.Source, doc.Title, doc.Checksum, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	return nil
}

func (store *PGVectorStore) GetDocumentByID(ctx context.Context, documentID string) (*Document, error) {
	var doc Document
	var metadataJSON []byte

	err := store.DB.QueryRowContext(ctx, `
		SELECT id, source, title, checksum, metadata
		FROM rag_documents
		WHERE id = $1
	`, documentID).Scan(
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

	doc.Metadata, err = unmarshalJSON(metadataJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal document metadata: %w", err)
	}

	return &doc, nil
}

func (store *PGVectorStore) ListDocuments(ctx context.Context) ([]DocumentSummary, error) {
	rows, err := store.DB.QueryContext(ctx, `
		SELECT
			d.id,
			COALESCE(d.source, ''),
			COALESCE(d.title, ''),
			COALESCE(d.checksum, ''),
			COALESCE(d.metadata, '{}'::jsonb),
			d.created_at,
			COUNT(c.id) AS chunk_count
		FROM rag_documents d
		LEFT JOIN rag_chunks c ON c.document_id = d.id
		GROUP BY d.id, d.source, d.title, d.checksum, d.metadata, d.created_at
		ORDER BY d.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()

	var documents []DocumentSummary
	for rows.Next() {
		var doc DocumentSummary
		var metadataJSON []byte

		if err := rows.Scan(
			&doc.ID,
			&doc.Source,
			&doc.Title,
			&doc.Checksum,
			&metadataJSON,
			&doc.CreatedAt,
			&doc.ChunkCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}

		doc.Metadata, err = unmarshalJSON(metadataJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal document metadata: %w", err)
		}
		documents = append(documents, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate documents: %w", err)
	}

	return documents, nil
}

func (store *PGVectorStore) DeleteDocument(ctx context.Context, documentID string) error {
	if _, err := store.DB.ExecContext(ctx, `DELETE FROM rag_documents WHERE id = $1`, documentID); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}
