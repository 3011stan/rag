package api

import (
	"context"
	"encoding/json"
	"net/http"

	demoseed "github.com/stan/Projects/studies/rag/internal/demo/seed"
	"github.com/stan/Projects/studies/rag/internal/logging"
)

func (srv *APIServer) SeedDemoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if srv.cfg.AdminToken == "" {
			respondError(w, "admin endpoint is disabled", http.StatusNotFound, nil)
			return
		}
		if !srv.isAuthorizedAdminRequest(r) {
			respondError(w, "unauthorized", http.StatusUnauthorized, nil)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), seedTimeout)
		defer cancel()

		seeded, err := demoseed.Directory(ctx, demoseed.DefaultDocsDir, srv.pipeline)
		if err != nil {
			logging.FromContext(ctx).Error().Err(err).Msg("demo seed failed")
			respondError(w, "failed to seed demo documents", http.StatusInternalServerError, nil)
			return
		}

		logging.FromContext(ctx).Info().
			Int("seeded", seeded).
			Msg("demo documents seeded")

		w.Header().Set(contentTypeHeader, contentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SeedDemoResponse{
			Seeded: seeded,
			Status: "success",
		})
	}
}
