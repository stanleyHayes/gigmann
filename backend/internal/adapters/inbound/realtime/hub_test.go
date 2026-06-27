package realtime //nolint:testpackage // white-box: exercises the unexported hub internals

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/token"
	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/user"
)

func TestNotifyBroadcastsToClients(t *testing.T) {
	h := New()
	c := &client{send: make(chan string, 1)}
	h.add(c)
	h.Notify("brief.refreshed")
	select {
	case ev := <-c.send:
		assert.Equal(t, "brief.refreshed", ev)
	default:
		t.Fatal("expected the event to be delivered")
	}
}

func TestNotifyDropsWhenBufferFull(t *testing.T) {
	h := New()
	h.add(&client{send: make(chan string)}) // unbuffered, no reader
	done := make(chan struct{})
	go func() { h.Notify("x"); close(done) }()
	select {
	case <-done: // did not block
	case <-time.After(time.Second):
		t.Fatal("Notify blocked on a full client buffer")
	}
}

func TestHandlerRejectsBadToken(t *testing.T) {
	tokens := token.New([]byte("secret"), time.Hour)
	rec := httptest.NewRecorder()
	New().Handler(tokens, nil)(rec, httptest.NewRequest(http.MethodGet, "/api/v1/ws?token=bad", nil))
	assert.Equal(t, 401, rec.Code)
}

func TestHandlerRejectsMissingToken(t *testing.T) {
	tokens := token.New([]byte("secret"), time.Hour)
	// a valid token still can't upgrade via httptest (no hijack) but must pass auth;
	// we only assert the auth gate here with a missing token.
	_ = auth.Principal{}
	_ = user.RoleExecutive
	rec := httptest.NewRecorder()
	New().Handler(tokens, []string{"http://localhost:5173"})(rec, httptest.NewRequest(http.MethodGet, "/api/v1/ws", nil))
	assert.Equal(t, 401, rec.Code)
}

func TestOriginPatternsStripsScheme(t *testing.T) {
	got := originPatterns([]string{"https://app.example.com", "localhost:5173"})
	require.Equal(t, []string{"app.example.com", "localhost:5173"}, got)
}
