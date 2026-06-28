package httpapi

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	sentryhttp "github.com/getsentry/sentry-go/http"
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
	"X-Content-Type-Options":       "nosniff",
	"X-Frame-Options":              "DENY",
	"Referrer-Policy":              "no-referrer",
	"Cross-Origin-Opener-Policy":   "same-origin",
	"Cross-Origin-Resource-Policy": "same-origin",
	// The API serves JSON only, so it never legitimately loads any resource.
	"Content-Security-Policy": "default-src 'none'; frame-ancestors 'none'; base-uri 'none'",
}

// securityHeaders sets conservative response security headers on every response.
// sentryMiddleware reports panics to Sentry (no-op when Sentry is not configured),
// then repanics so the chi Recoverer still returns a 500. Registered after
// Recoverer so it unwinds first.
func sentryMiddleware() func(http.Handler) http.Handler {
	h := sentryhttp.New(sentryhttp.Options{Repanic: true})
	return func(next http.Handler) http.Handler { return h.Handle(next) }
}

func securityHeaders(hsts bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for k, v := range securityHeaderValues {
				w.Header().Set(k, v)
			}
			// HSTS only over TLS (production); on plain-HTTP localhost it would pin
			// the dev host to HTTPS for two years.
			if hsts {
				w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
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
				h.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
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

// readyHandler reports readiness; when a dependency check is supplied (e.g. a DB
// ping) and it fails, it returns 503 so the orchestrator holds traffic.
func readyHandler(check func(context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if check != nil {
			if err := check(r.Context()); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte(`{"status":"unavailable"}`))
				return
			}
		}
		writeReady(w, r)
	}
}

// rateLimiter is a simple in-memory fixed-window per-key request limiter. It is
// per-process (fine for the demo); a clustered deploy would back this with Redis.
type rateLimiter struct {
	mu        sync.Mutex
	limit     int
	window    time.Duration
	now       func() time.Time
	counts    map[string]*windowCount
	lastSweep time.Time
}

type windowCount struct {
	count   int
	resetAt time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter { //nolint:unparam // window is a deliberate per-limiter knob (auth vs ask)
	return &rateLimiter{limit: limit, window: window, now: time.Now, counts: map[string]*windowCount{}}
}

// rateLimitPrincipal limits the given paths per authenticated principal (falling
// back to client IP for unauthenticated calls). It must run AFTER authMiddleware
// so the principal is in context. Used to bound AI cost/abuse on the Ask endpoint.
func rateLimitPrincipal(l *rateLimiter, prefixes []string, trustProxy bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if pathLimited(r.URL.Path, prefixes) {
				key := "ip:" + clientIP(r, trustProxy)
				if p, ok := principalFrom(r.Context()); ok && p.UserID != "" {
					key = "user:" + p.UserID
				}
				if !l.allow(key) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusTooManyRequests)
					_, _ = w.Write([]byte(`{"error":"rate_limited"}`))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (l *rateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.sweep(now)
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

// sweep evicts expired windows, at most once per window, so the counts map cannot
// grow unbounded with one entry per distinct client/principal seen over time.
// Caller must hold l.mu.
func (l *rateLimiter) sweep(now time.Time) {
	if !now.After(l.lastSweep.Add(l.window)) {
		return
	}
	for k, wc := range l.counts {
		if now.After(wc.resetAt) {
			delete(l.counts, k)
		}
	}
	l.lastSweep = now
}

// clientIP extracts the caller IP for rate limiting. X-Forwarded-For is honoured
// ONLY when trustProxy is set (the deployment is behind a trusted proxy such as
// Render); otherwise a client could spoof X-Forwarded-For to bypass the per-IP
// limit. When trusted, the rightmost entry is used — it is the address the proxy
// actually observed, whereas any leftmost entries are client-supplied.
func clientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if ip := strings.TrimSpace(parts[len(parts)-1]); ip != "" {
				return ip
			}
		}
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
func rateLimit(l *rateLimiter, prefixes []string, trustProxy bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if pathLimited(r.URL.Path, prefixes) && !l.allow(clientIP(r, trustProxy)) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"rate_limited"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
