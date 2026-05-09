package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/stan/Projects/studies/rag/internal/logging"
	"github.com/stan/Projects/studies/rag/internal/rag/loader"
)

func (srv *APIServer) IngestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !srv.cfg.PublicUploadEnabled && !srv.isAuthorizedAdminRequest(r) {
			if srv.cfg.AdminToken == "" {
				respondError(w, "ingest endpoint is disabled", http.StatusNotFound, nil)
				return
			}
			respondError(w, "unauthorized", http.StatusUnauthorized, nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), ingestTimeout)
		defer cancel()
		r.Body = http.MaxBytesReader(w, r.Body, srv.cfg.MaxUploadBytes)

		logger := logging.FromContext(ctx)
		uploaded, err := srv.parseAndValidateFile(r)
		if err != nil {
			logger.Warn().Err(err).Msg("validation failed")
			respondError(w, err.Error(), http.StatusBadRequest, nil)
			return
		}

		logger.Info().
			Str("filename", uploaded.Name).
			Str("content_type", uploaded.ContentType).
			Int("size_bytes", len(uploaded.Data)).
			Msg("processing document")

		result, err := srv.pipeline.Ingest(ctx, loader.Source{
			Name:        uploaded.Name,
			ContentType: uploaded.ContentType,
			Data:        uploaded.Data,
		})
		if err != nil {
			if errors.Is(err, loader.ErrUnsupportedType) || errors.Is(err, loader.ErrInvalidDocument) {
				logger.Warn().Err(err).Msg("document validation failed")
				respondError(w, err.Error(), http.StatusBadRequest, nil)
				return
			}

			logger.Error().Err(err).Msg("ingest pipeline failed")
			respondError(w, "failed to ingest document", http.StatusInternalServerError, nil)
			return
		}

		w.Header().Set(contentTypeHeader, contentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(IngestResponse{
			DocumentID: result.DocumentID,
			ChunkCount: result.ChunkCount,
			Status:     "success",
			Message:    fmt.Sprintf("Successfully ingested document: %s", uploaded.Name),
		})
	}
}
