package httpapi //nolint:testpackage // white-box: exercises the unexported middleware

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	h := rateLimit(rl, []string{"/api/v1/auth/login"})(ok)

	for i, want := range []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = "1.2.3.4:5555"
		h.ServeHTTP(rec, req)
		assert.Equal(t, want, rec.Code, "request %d", i)
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

func TestClientIP(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "1.2.3.4:55"
	assert.Equal(t, "1.2.3.4", clientIP(r))
	r.Header.Set("X-Forwarded-For", "8.8.8.8, 1.1.1.1")
	assert.Equal(t, "8.8.8.8", clientIP(r))
}
