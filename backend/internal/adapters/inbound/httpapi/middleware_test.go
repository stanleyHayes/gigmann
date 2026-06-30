package httpapi //nolint:testpackage // white-box: exercises the unexported middleware

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/core/auth"
)

func TestRequestLoggerLogs(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	h := requestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/widgets", nil))

	out := buf.String()
	assert.Contains(t, out, "http_request")
	assert.Contains(t, out, `"method":"GET"`)
	assert.Contains(t, out, `"path":"/widgets"`)
	assert.Contains(t, out, `"status":201`)
}

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	h := corsMiddleware([]string{"https://app.example"})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/x", nil)
	req.Header.Set("Origin", "https://app.example")
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "https://app.example", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSRejectsUnknownOrigin(t *testing.T) {
	h := corsMiddleware([]string{"https://app.example"})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "https://evil.example")
	h.ServeHTTP(rec, req)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestSecurityHeaders(t *testing.T) {
	h := securityHeaders(true)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	assert.Equal(t, "max-age=63072000; includeSubDomains", rec.Header().Get("Strict-Transport-Security"))
	assert.Contains(t, rec.Header().Get("Content-Security-Policy"), "default-src 'none'")
	assert.Equal(t, "same-origin", rec.Header().Get("Cross-Origin-Resource-Policy"))
	assert.Equal(t, "?1", rec.Header().Get("Origin-Agent-Cluster"))
	assert.Equal(t, "off", rec.Header().Get("X-DNS-Prefetch-Control"))
	assert.Equal(t, "none", rec.Header().Get("X-Permitted-Cross-Domain-Policies"))
	assert.Contains(t, rec.Header().Get("Permissions-Policy"), "geolocation=()")
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
}

func TestSecurityHeadersNoHSTSWhenDisabled(t *testing.T) {
	h := securityHeaders(false)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	assert.Empty(t, rec.Header().Get("Strict-Transport-Security"))
}

func TestReadyHandler(t *testing.T) {
	// No check supplied → ready.
	rec := httptest.NewRecorder()
	readyHandler(nil)(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ready")

	// Failing dependency check → 503.
	rec2 := httptest.NewRecorder()
	readyHandler(func(context.Context) error { return errors.New("db down") })(rec2, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	assert.Equal(t, http.StatusServiceUnavailable, rec2.Code)
}

func TestRateLimitBlocksAfterLimit(t *testing.T) {
	rl := newRateLimiter(2, time.Minute)
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := rateLimit(rl, []string{"/api/v1/auth/login"}, false)(ok)

	for i, want := range []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = "1.2.3.4:5555"
		h.ServeHTTP(rec, req)
		assert.Equal(t, want, rec.Code, "request %d", i)
		if want == http.StatusTooManyRequests {
			assert.Equal(t, "60", rec.Header().Get("Retry-After"))
		}
	}

	// a different IP is unaffected
	other := httptest.NewRecorder()
	oreq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	oreq.RemoteAddr = "9.9.9.9:1"
	h.ServeHTTP(other, oreq)
	assert.Equal(t, http.StatusOK, other.Code)

	// a non-limited path is never throttled, even from the blocked IP
	free := httptest.NewRecorder()
	freq := httptest.NewRequest(http.MethodGet, "/api/v1/brief", nil)
	freq.RemoteAddr = "1.2.3.4:5555"
	h.ServeHTTP(free, freq)
	assert.Equal(t, http.StatusOK, free.Code)
}

func TestRateLimiterEvictsExpiredEntries(t *testing.T) {
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	cur := base
	rl := newRateLimiter(5, time.Minute)
	rl.now = func() time.Time { return cur }

	for i := range 10 {
		ok, _ := rl.allow("key-" + strconv.Itoa(i))
		require.True(t, ok)
	}
	rl.mu.Lock()
	filled := len(rl.counts)
	rl.mu.Unlock()
	require.Equal(t, 10, filled, "one window entry per distinct key")

	// Past the window + the once-per-window sweep interval: the next call evicts
	// every now-expired window so the map cannot grow unbounded.
	cur = base.Add(3 * time.Minute)
	ok, _ := rl.allow("fresh")
	require.True(t, ok)
	rl.mu.Lock()
	after := len(rl.counts)
	rl.mu.Unlock()
	assert.Equal(t, 1, after, "expired windows evicted; only the fresh key remains")
}

func TestClientIP(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "1.2.3.4:55"
	// Without a trusted proxy, X-Forwarded-For is ignored (it is client-spoofable).
	assert.Equal(t, "1.2.3.4", clientIP(r, false))
	r.Header.Set("X-Forwarded-For", "8.8.8.8, 1.1.1.1")
	assert.Equal(t, "1.2.3.4", clientIP(r, false), "XFF ignored when not behind a trusted proxy")
	// Behind a trusted proxy, the rightmost (proxy-observed) entry is used; the
	// spoofable leftmost 8.8.8.8 is not trusted.
	assert.Equal(t, "1.1.1.1", clientIP(r, true))
}

func TestRateLimitPrincipal(t *testing.T) {
	rl := newRateLimiter(2, time.Minute)
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := rateLimitPrincipal(rl, []string{"/api/v1/ask"}, false)(ok)

	do := func(userID string) int {
		ctx := withPrincipal(context.Background(), auth.Principal{UserID: userID})
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ask", nil).WithContext(ctx)
		req.RemoteAddr = "9.9.9.9:1" // same IP for all → proves keying is per-principal
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	assert.Equal(t, http.StatusOK, do("u1"))
	assert.Equal(t, http.StatusOK, do("u1"))
	assert.Equal(t, http.StatusTooManyRequests, do("u1"), "3rd call from u1 exceeds limit 2")
	assert.Equal(t, http.StatusOK, do("u2"), "a different principal is unaffected")
}

func TestRateLimitPrincipalIgnoresOtherPaths(t *testing.T) {
	rl := newRateLimiter(1, time.Minute)
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := rateLimitPrincipal(rl, []string{"/api/v1/ask"}, false)(ok)
	for range 3 {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/brief", nil)
		h.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code) // unlimited path
	}
}
