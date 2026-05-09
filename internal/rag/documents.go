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

func (store *PGVectorStore) DeleteDocument(ctx context.Context, documentID string) error {
	if _, err := store.DB.ExecContext(ctx, `DELETE FROM rag_documents WHERE id = $1`, documentID); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}
