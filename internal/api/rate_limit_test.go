package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stan/Projects/studies/rag/internal/config"
)

func TestRateLimitMiddlewareAllowsRequestsUnderLimit(t *testing.T) {
	server := &APIServer{cfg: &config.Config{
		RateLimitEnabled:    true,
		RateLimitRequests:   2,
		RateLimitWindowSecs: 60,
	}}
	handler := server.RateLimitMiddleware()(okHandler())

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/rag/ask", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected request %d to pass, got %d", i+1, rec.Code)
		}
	}
}

func TestRateLimitMiddlewareBlocksAfterLimit(t *testing.T) {
	server := &APIServer{cfg: &config.Config{
		RateLimitEnabled:    true,
		RateLimitRequests:   1,
		RateLimitWindowSecs: 60,
	}}
	handler := server.RateLimitMiddleware()(okHandler())
	req := httptest.NewRequest(http.MethodPost, "/rag/ask", nil)
	req.RemoteAddr = "203.0.113.10:1234"

	handler.ServeHTTP(httptest.NewRecorder(), req)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
	if rec.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Fatalf("expected zero remaining requests, got %q", rec.Header().Get("X-RateLimit-Remaining"))
	}
}

func TestRateLimitMiddlewareSeparatesClientIPs(t *testing.T) {
	server := &APIServer{cfg: &config.Config{
		RateLimitEnabled:    true,
		RateLimitRequests:   1,
		RateLimitWindowSecs: 60,
	}}
	handler := server.RateLimitMiddleware()(okHandler())

	first := httptest.NewRequest(http.MethodPost, "/rag/ask", nil)
	first.RemoteAddr = "203.0.113.10:1234"
	handler.ServeHTTP(httptest.NewRecorder(), first)

	second := httptest.NewRequest(http.MethodPost, "/rag/ask", nil)
	second.RemoteAddr = "203.0.113.11:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, second)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected different client IP to pass, got %d", rec.Code)
	}
}

func TestRateLimitMiddlewareUsesForwardedFor(t *testing.T) {
	server := &APIServer{cfg: &config.Config{
		RateLimitEnabled:    true,
		RateLimitRequests:   1,
		RateLimitWindowSecs: 60,
	}}
	handler := server.RateLimitMiddleware()(okHandler())

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/rag/ask", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		req.Header.Set("X-Forwarded-For", "203.0.113.10, 10.0.0.1")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if i == 1 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("expected second forwarded request to be blocked, got %d", rec.Code)
		}
	}
}

func TestRateLimitMiddlewareIgnoresInvalidForwardedFor(t *testing.T) {
	server := &APIServer{cfg: &config.Config{
		RateLimitEnabled:    true,
		RateLimitRequests:   1,
		RateLimitWindowSecs: 60,
	}}
	handler := server.RateLimitMiddleware()(okHandler())

	req := httptest.NewRequest(http.MethodPost, "/rag/ask", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	req.Header.Set("X-Forwarded-For", "not-an-ip")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected fallback RemoteAddr IP to be rate limited, got %d", rec.Code)
	}
}

func TestRateLimitMiddlewareDisabledAllowsAllRequests(t *testing.T) {
	server := &APIServer{cfg: &config.Config{
		RateLimitEnabled:    false,
		RateLimitRequests:   1,
		RateLimitWindowSecs: 60,
	}}
	handler := server.RateLimitMiddleware()(okHandler())

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/rag/ask", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected disabled limiter to pass request %d, got %d", i+1, rec.Code)
		}
	}
}

func TestRateLimiterResetsAfterWindow(t *testing.T) {
	limiter := newRateLimiter(1, time.Minute)
	now := time.Date(2026, 6, 11, 10, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }

	if result := limiter.allow("client"); !result.Allowed {
		t.Fatal("expected first request to pass")
	}
	if result := limiter.allow("client"); result.Allowed {
		t.Fatal("expected second request to be blocked")
	}

	now = now.Add(time.Minute)
	if result := limiter.allow("client"); !result.Allowed {
		t.Fatal("expected request to pass after window reset")
	}
}

func TestRateLimiterCleanupIsPeriodic(t *testing.T) {
	limiter := newRateLimiter(1, time.Minute)
	now := time.Date(2026, 6, 11, 10, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }

	limiter.allow("first")
	now = now.Add(30 * time.Second)
	limiter.allow("second")

	if _, exists := limiter.clients["first"]; !exists {
		t.Fatal("expected first client to remain before cleanup interval")
	}

	now = now.Add(31 * time.Second)
	limiter.allow("third")

	if _, exists := limiter.clients["first"]; exists {
		t.Fatal("expected expired client to be cleaned up")
	}
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
