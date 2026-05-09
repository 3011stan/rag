package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stan/Projects/studies/rag/internal/logging"
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
		fileData, fileName, err := srv.parseAndValidateFile(r)
		if err != nil {
			logger.Warn().Err(err).Msg("validation failed")
			respondError(w, err.Error(), http.StatusBadRequest, nil)
			return
		}

		logger.Info().
			Str("filename", fileName).
			Int("size_bytes", len(fileData)).
			Msg("processing PDF")

		result, err := srv.pipeline.IngestPDF(ctx, fileData, fileName)
		if err != nil {
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
			Message:    fmt.Sprintf("Successfully ingested PDF: %s", fileName),
		})
	}
}
