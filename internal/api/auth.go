package api

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

func (srv *APIServer) isAuthorizedAdminRequest(r *http.Request) bool {
	token := r.Header.Get("X-Admin-Token")
	if token == "" {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return false
		}
		token = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	}
	if token == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(srv.cfg.AdminToken)) == 1
}
