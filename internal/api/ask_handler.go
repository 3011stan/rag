package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/stan/Projects/studies/rag/internal/logging"
)

func (srv *APIServer) AskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), askTimeout)
		defer cancel()
		r.Body = http.MaxBytesReader(w, r.Body, maxAskBodySize)

		logger := logging.FromContext(ctx)
		req, err := srv.parseAndValidateQuestion(r)
		if err != nil {
			logger.Warn().Err(err).Msg("validation failed")
			respondError(w, err.Error(), http.StatusBadRequest, nil)
			return
		}

		preview := req.Question
		if len(preview) > 50 {
			preview = preview[:50]
		}
		logger.Info().
			Str("question_preview", preview).
			Int("top_k", req.TopK).
			Msg("processing question")

		result, err := srv.pipeline.Ask(ctx, req.Question, req.TopK)
		if err != nil {
			logger.Error().Err(err).Msg("answer pipeline failed")
			respondError(w, "failed to answer question", http.StatusInternalServerError, nil)
			return
		}

		resp := AskResponse{
			Answer:  result.Answer,
			Sources: make([]SourceInfo, len(result.Sources)),
		}
		for i, source := range result.Sources {
			resp.Sources[i] = SourceInfo{
				DocumentID: source.DocumentID,
				ChunkIndex: source.ChunkIndex,
				Score:      source.Score,
				Preview:    source.Preview,
			}
		}

		logger.Info().
			Int("source_count", len(resp.Sources)).
			Msg("answer generated successfully")

		w.Header().Set(contentTypeHeader, contentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}
