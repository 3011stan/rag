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
	mu      sync.Mutex
	clients map[string]rateLimitWindow
	limit   int
	window  time.Duration
	now     func() time.Time
}

type rateLimitWindow struct {
	start time.Time
	count int
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		clients: make(map[string]rateLimitWindow),
		limit:   limit,
		window:  window,
		now:     time.Now,
	}
}

func (l *rateLimiter) allow(key string) (bool, int, time.Duration) {
	if l == nil || l.limit <= 0 || l.window <= 0 {
		return true, 0, 0
	}

	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()

	for client, window := range l.clients {
		if now.Sub(window.start) >= l.window {
			delete(l.clients, client)
		}
	}

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
		return false, remaining, l.window - now.Sub(window.start)
	}
	return true, remaining, 0
}

func (srv *APIServer) RateLimitMiddleware() func(http.Handler) http.Handler {
	if srv == nil || srv.cfg == nil || !srv.cfg.RateLimitEnabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	limiter := newRateLimiter(srv.cfg.RateLimitRequests, time.Duration(srv.cfg.RateLimitWindowSecs)*time.Second)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, remaining, retryAfter := limiter.allow(clientIP(r))
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limiter.limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			if !allowed {
				if retryAfter < time.Second {
					retryAfter = time.Second
				}
				w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
				respondError(w, "rate limit exceeded", http.StatusTooManyRequests, nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		if ip := strings.TrimSpace(strings.Split(forwardedFor, ",")[0]); ip != "" {
			return ip
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
