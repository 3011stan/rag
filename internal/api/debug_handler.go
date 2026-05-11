package api

import (
	"encoding/json"
	"net/http"
)

func (srv *APIServer) DebugMetadataHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !srv.isAuthorizedAdminRequest(r) {
			respondError(w, "unauthorized", http.StatusUnauthorized, nil)
			return
		}

		w.Header().Set(contentTypeHeader, contentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(DebugMetadataResponse{
			Environment:         srv.cfg.Env,
			AIProvider:          srv.cfg.ResolvedAIProvider(),
			EmbeddingModel:      srv.cfg.EmbeddingModel,
			LLMModel:            srv.cfg.LLMModel,
			EmbeddingDimensions: srv.cfg.EmbeddingDimensions,
			TopK:                srv.cfg.TopK,
			ChunkTokens:         srv.cfg.ChunkTokens,
			OverlapTokens:       srv.cfg.OverlapTokens,
			PublicUploadEnabled: srv.cfg.PublicUploadEnabled,
		})
	}
}
