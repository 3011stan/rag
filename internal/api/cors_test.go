package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stan/Projects/studies/rag/internal/config"
)

func TestCORSMiddlewareAllowsConfiguredOriginPreflight(t *testing.T) {
	server := &APIServer{
		cfg: &config.Config{
			CORSAllowedOrigins: []string{"https://rag-lab.vercel.app"},
		},
	}

	handler := server.CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/rag/ask", nil)
	req.Header.Set("Origin", "https://rag-lab.vercel.app")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://rag-lab.vercel.app" {
		t.Fatalf("expected configured origin, got %q", got)
	}
}

func TestCORSMiddlewareDoesNotAllowUnconfiguredOrigin(t *testing.T) {
	server := &APIServer{
		cfg: &config.Config{
			CORSAllowedOrigins: []string{"https://rag-lab.vercel.app"},
		},
	}

	handler := server.CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/rag/ask", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no allow-origin header, got %q", got)
	}
}
