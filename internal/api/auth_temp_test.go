package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stan/Projects/studies/rag/internal/config"
)

func TestTemporaryTokenHandlerIssuesScopedToken(t *testing.T) {
	server := &APIServer{
		cfg: &config.Config{TemporaryTokenSecret: "test-temp-secret"},
	}

	req := httptest.NewRequest("POST", "/auth/temp-token", bytes.NewBufferString(`{"email":"User@Example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.TemporaryTokenHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp TemporaryTokenResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected token")
	}
	if len(resp.Scopes) != 2 {
		t.Fatalf("expected two scopes, got %d", len(resp.Scopes))
	}

	claims, err := server.verifyTemporaryToken(resp.Token)
	if err != nil {
		t.Fatalf("expected token to verify: %v", err)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", claims.Email)
	}
	if !server.isAuthorizedTemporaryRequest(requestWithBearer(resp.Token), tempTokenScopeUpload) {
		t.Fatal("expected token to authorize upload scope")
	}
	if server.isAuthorizedAdminRequest(requestWithBearer(resp.Token)) {
		t.Fatal("temporary token must not authorize admin requests")
	}
}

func TestTemporaryTokenRejectsMissingSecret(t *testing.T) {
	server := &APIServer{cfg: &config.Config{}}

	req := httptest.NewRequest("POST", "/auth/temp-token", bytes.NewBufferString(`{"email":"user@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.TemporaryTokenHandler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestVerifyTemporaryTokenRejectsExpiredToken(t *testing.T) {
	server := &APIServer{cfg: &config.Config{TemporaryTokenSecret: "test-temp-secret"}}
	claims := temporaryTokenClaims{
		Email:     "user@example.com",
		Scopes:    []string{tempTokenScopeUpload},
		ExpiresAt: time.Now().Add(-time.Minute).Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("failed to marshal claims: %v", err)
	}
	encoded := base64Encode(payload)
	token := encoded + "." + signTemporaryToken(encoded, "test-temp-secret")

	if _, err := server.verifyTemporaryToken(token); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}

func requestWithBearer(token string) *http.Request {
	req := httptest.NewRequest("GET", "/documents", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func base64Encode(payload []byte) string {
	return base64.RawURLEncoding.EncodeToString(payload)
}
