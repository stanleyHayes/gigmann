package httpapi

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

const corsMaxAge = "600"

// requestLogger logs one structured line per request (no PII — method, path,
// status, duration, request id only).
func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)
			logger.LogAttrs(r.Context(), slog.LevelInfo, "http_request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)
		})
	}
}

var securityHeaderValues = map[string]string{
	"X-Content-Type-Options":     "nosniff",
	"X-Frame-Options":            "DENY",
	"Referrer-Policy":            "no-referrer",
	"Cross-Origin-Opener-Policy": "same-origin",
}

// securityHeaders sets conservative response security headers on every response.
func securityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for k, v := range securityHeaderValues {
				w.Header().Set(k, v)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// corsMiddleware allows the configured origins (allow-list) and answers
// preflight requests; unknown origins receive no CORS headers.
func corsMiddleware(origins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(origins))
	for _, o := range origins {
		allowed[o] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if origin := r.Header.Get("Origin"); origin != "" && allowed[origin] {
				h := w.Header()
				h.Set("Access-Control-Allow-Origin", origin)
				h.Add("Vary", "Origin")
				h.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				h.Set("Access-Control-Allow-Credentials", "true")
				h.Set("Access-Control-Max-Age", corsMaxAge)
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// writeReady is the readiness probe (the in-memory demo is always ready).
func writeReady(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ready"}`))
}

// rateLimiter is a simple in-memory fixed-window per-key request limiter. It is
// per-process (fine for the demo); a clustered deploy would back this with Redis.
type rateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	now    func() time.Time
	counts map[string]*windowCount
}

type windowCount struct {
	count   int
	resetAt time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{limit: limit, window: window, now: time.Now, counts: map[string]*windowCount{}}
}

func (l *rateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	wc := l.counts[key]
	if wc == nil || now.After(wc.resetAt) {
		l.counts[key] = &windowCount{count: 1, resetAt: now.Add(l.window)}
		return true
	}
	if wc.count >= l.limit {
		return false
	}
	wc.count++
	return true
}

// clientIP extracts the caller IP, honouring X-Forwarded-For behind a proxy.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		first, _, _ := strings.Cut(xff, ",")
		return strings.TrimSpace(first)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func pathLimited(path string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// rateLimit throttles per-IP requests to the given path prefixes (brute-force
// protection for auth); other paths pass freely.
func rateLimit(l *rateLimiter, prefixes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if pathLimited(r.URL.Path, prefixes) && !l.allow(clientIP(r)) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"rate_limited"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
