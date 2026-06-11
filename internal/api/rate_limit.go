package api

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type rateLimiter struct {
	mu          sync.Mutex
	clients     map[string]rateLimitWindow
	limit       int
	window      time.Duration
	lastCleanup time.Time
	now         func() time.Time
}

type rateLimitWindow struct {
	start time.Time
	count int
}

type rateLimitResult struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		clients: make(map[string]rateLimitWindow),
		limit:   limit,
		window:  window,
		now:     time.Now,
	}
}

func (l *rateLimiter) allow(key string) rateLimitResult {
	if l == nil || l.limit <= 0 || l.window <= 0 {
		return rateLimitResult{Allowed: true}
	}

	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()

	l.cleanupExpired(now)

	window := l.clients[key]
	if window.start.IsZero() || now.Sub(window.start) >= l.window {
		window = rateLimitWindow{start: now}
	}
	window.count++
	l.clients[key] = window

	remaining := l.limit - window.count
	if remaining < 0 {
		remaining = 0
	}
	if window.count > l.limit {
		return rateLimitResult{
			Allowed:    false,
			Remaining:  remaining,
			RetryAfter: l.window - now.Sub(window.start),
		}
	}
	return rateLimitResult{Allowed: true, Remaining: remaining}
}

func (l *rateLimiter) cleanupExpired(now time.Time) {
	if !l.lastCleanup.IsZero() && now.Sub(l.lastCleanup) < l.window {
		return
	}
	for client, window := range l.clients {
		if now.Sub(window.start) >= l.window {
			delete(l.clients, client)
		}
	}
	l.lastCleanup = now
}

func (srv *APIServer) RateLimitMiddleware() func(http.Handler) http.Handler {
	if !srv.rateLimitEnabled() {
		return noopMiddleware
	}

	limiter := newRateLimiter(srv.cfg.RateLimitRequests, time.Duration(srv.cfg.RateLimitWindowSecs)*time.Second)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			result := limiter.allow(clientIP(r))
			writeRateLimitHeaders(w, limiter.limit, result)
			if !result.Allowed {
				writeRateLimitExceeded(w, result.RetryAfter)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (srv *APIServer) rateLimitEnabled() bool {
	return srv != nil && srv.cfg != nil && srv.cfg.RateLimitEnabled
}

func noopMiddleware(next http.Handler) http.Handler {
	return next
}

func writeRateLimitHeaders(w http.ResponseWriter, limit int, result rateLimitResult) {
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
}

func writeRateLimitExceeded(w http.ResponseWriter, retryAfter time.Duration) {
	if retryAfter < time.Second {
		retryAfter = time.Second
	}
	w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
	respondError(w, "rate limit exceeded", http.StatusTooManyRequests, nil)
}

func clientIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		if ip := parseClientIP(strings.Split(forwardedFor, ",")[0]); ip != "" {
			return ip
		}
	}
	if realIP := parseClientIP(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		if ip := parseClientIP(host); ip != "" {
			return ip
		}
	}
	if ip := parseClientIP(r.RemoteAddr); ip != "" {
		return ip
	}
	return "unknown"
}

func parseClientIP(value string) string {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil {
		return ""
	}
	return ip.String()
}
