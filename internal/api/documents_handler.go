package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stan/Projects/studies/rag/internal/logging"
)

func (srv *APIServer) ListDocumentsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !srv.isAuthorizedAdminRequest(r) && !srv.isAuthorizedTemporaryRequest(r, tempTokenScopeListDocuments) {
			respondError(w, "unauthorized", http.StatusUnauthorized, nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), documentsTimeout)
		defer cancel()

		documents, err := srv.vs.ListDocuments(ctx)
		if err != nil {
			logging.FromContext(ctx).Error().Err(err).Msg("failed to list documents")
			respondError(w, "failed to list documents", http.StatusInternalServerError, nil)
			return
		}

		resp := DocumentsResponse{Documents: make([]DocumentInfo, len(documents))}
		for i, doc := range documents {
			resp.Documents[i] = DocumentInfo{
				ID:         doc.ID,
				Source:     doc.Source,
				Title:      doc.Title,
				Checksum:   doc.Checksum,
				ChunkCount: doc.ChunkCount,
				Metadata:   doc.Metadata,
				CreatedAt:  doc.CreatedAt.Format(time.RFC3339),
			}
		}

		w.Header().Set(contentTypeHeader, contentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func (srv *APIServer) DeleteDocumentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !srv.isAuthorizedAdminRequest(r) {
			respondError(w, "unauthorized", http.StatusUnauthorized, nil)
			return
		}

		documentID := chi.URLParam(r, "id")
		if documentID == "" {
			respondError(w, "document id is required", http.StatusBadRequest, nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), documentsTimeout)
		defer cancel()

		if err := srv.vs.DeleteDocument(ctx, documentID); err != nil {
			logging.FromContext(ctx).Error().Err(err).Str("document_id", documentID).Msg("failed to delete document")
			respondError(w, "failed to delete document", http.StatusInternalServerError, nil)
			return
		}

		w.Header().Set(contentTypeHeader, contentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(DeleteDocumentResponse{
			DocumentID: documentID,
			Status:     "deleted",
		})
	}
}
