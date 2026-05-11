package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	tempTokenScopeUpload        = "upload"
	tempTokenScopeListDocuments = "documents:list"
	tempTokenDuration           = 30 * time.Minute
)

type temporaryTokenClaims struct {
	Email     string   `json:"email"`
	Scopes    []string `json:"scopes"`
	ExpiresAt int64    `json:"expires_at"`
}

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

func (srv *APIServer) isAuthorizedTemporaryRequest(r *http.Request, requiredScope string) bool {
	claims, ok := srv.temporaryTokenClaimsFromRequest(r)
	if !ok {
		return false
	}
	for _, scope := range claims.Scopes {
		if scope == requiredScope {
			return true
		}
	}
	return false
}

func (srv *APIServer) temporaryTokenClaimsFromRequest(r *http.Request) (temporaryTokenClaims, bool) {
	token := bearerTokenFromRequest(r)
	if token == "" {
		return temporaryTokenClaims{}, false
	}
	claims, err := srv.verifyTemporaryToken(token)
	if err != nil {
		return temporaryTokenClaims{}, false
	}
	return claims, true
}

func (srv *APIServer) createTemporaryToken(email string) (string, time.Time, error) {
	secret := srv.temporaryTokenSecret()
	if secret == "" {
		return "", time.Time{}, fmt.Errorf("temporary token endpoint is disabled")
	}

	expiresAt := time.Now().UTC().Add(tempTokenDuration)
	claims := temporaryTokenClaims{
		Email: strings.ToLower(strings.TrimSpace(email)),
		Scopes: []string{
			tempTokenScopeUpload,
			tempTokenScopeListDocuments,
		},
		ExpiresAt: expiresAt.Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to encode token claims: %w", err)
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature := signTemporaryToken(encodedPayload, secret)
	return encodedPayload + "." + signature, expiresAt, nil
}

func (srv *APIServer) verifyTemporaryToken(token string) (temporaryTokenClaims, error) {
	secret := srv.temporaryTokenSecret()
	if secret == "" {
		return temporaryTokenClaims{}, fmt.Errorf("temporary token auth is disabled")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return temporaryTokenClaims{}, fmt.Errorf("invalid token format")
	}

	expectedSignature := signTemporaryToken(parts[0], secret)
	if subtle.ConstantTimeCompare([]byte(parts[1]), []byte(expectedSignature)) != 1 {
		return temporaryTokenClaims{}, fmt.Errorf("invalid token signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return temporaryTokenClaims{}, fmt.Errorf("invalid token payload: %w", err)
	}

	var claims temporaryTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return temporaryTokenClaims{}, fmt.Errorf("invalid token claims: %w", err)
	}
	if claims.ExpiresAt <= time.Now().UTC().Unix() {
		return temporaryTokenClaims{}, fmt.Errorf("temporary token expired")
	}

	return claims, nil
}

func (srv *APIServer) temporaryTokenSecret() string {
	return srv.cfg.TemporaryTokenSecret
}

func bearerTokenFromRequest(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
}

func signTemporaryToken(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
