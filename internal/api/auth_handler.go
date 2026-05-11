package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (srv *APIServer) TemporaryTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, cancel := context.WithTimeout(r.Context(), authTimeout)
		defer cancel()

		req, err := parseTemporaryTokenRequest(r)
		if err != nil {
			respondError(w, err.Error(), http.StatusBadRequest, nil)
			return
		}

		token, expiresAt, err := srv.createTemporaryToken(req.Email)
		if err != nil {
			respondError(w, err.Error(), http.StatusNotFound, nil)
			return
		}

		w.Header().Set(contentTypeHeader, contentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TemporaryTokenResponse{
			Token:     token,
			ExpiresAt: expiresAt.Format(time.RFC3339),
			Scopes: []string{
				tempTokenScopeUpload,
				tempTokenScopeListDocuments,
			},
		})
	}
}

func parseTemporaryTokenRequest(r *http.Request) (TemporaryTokenRequest, error) {
	if !strings.HasPrefix(r.Header.Get(contentTypeHeader), contentTypeJSON) {
		return TemporaryTokenRequest{}, fmt.Errorf("content type must be application/json")
	}

	var req TemporaryTokenRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return TemporaryTokenRequest{}, fmt.Errorf("invalid request format: %v", err)
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		return TemporaryTokenRequest{}, fmt.Errorf("valid email is required")
	}

	return req, nil
}
