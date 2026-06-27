// Package realtime is an inbound adapter: a WebSocket hub that pushes events
// (e.g. "brief.refreshed") to connected clients. Single-instance (no Redis); it
// implements ports.Notifier so the application can broadcast without importing it.
package realtime

import (
	"net/http"
	"sync"

	"github.com/coder/websocket"

	"github.com/xcreativs/gigmann/internal/ports"
)

const sendBuffer = 16

type client struct{ send chan string }

// Hub fans named events out to all connected WebSocket clients.
type Hub struct {
	mu      sync.Mutex
	clients map[*client]struct{}
}

// New creates an empty hub.
func New() *Hub { return &Hub{clients: map[*client]struct{}{}} }

var _ ports.Notifier = (*Hub)(nil)

// Notify broadcasts an event to every connected client (dropping it for any
// client whose buffer is full, so one slow reader can't block the others).
func (h *Hub) Notify(event string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		select {
		case c.send <- event:
		default:
		}
	}
}

// Handler authenticates (token query param, since browsers can't set the Bearer
// header on a WebSocket) and upgrades the connection, then pushes events.
func (h *Hub) Handler(tokens ports.TokenService, allowedOrigins []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := tokens.Verify(r.URL.Query().Get("token")); err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: originPatterns(allowedOrigins)})
		if err != nil {
			return
		}
		defer func() { _ = conn.CloseNow() }()

		c := &client{send: make(chan string, sendBuffer)}
		h.add(c)
		defer h.remove(c)

		ctx := conn.CloseRead(r.Context()) // push-only; CloseRead handles pings/closes
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-c.send:
				if err := conn.Write(ctx, websocket.MessageText, []byte(event)); err != nil {
					return
				}
			}
		}
	}
}

func (h *Hub) add(c *client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) remove(c *client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

func originPatterns(origins []string) []string {
	out := make([]string, 0, len(origins))
	for _, o := range origins {
		// websocket.Accept matches the Host of the Origin URL; strip the scheme.
		if i := indexAfter(o, "://"); i >= 0 {
			out = append(out, o[i:])
		} else {
			out = append(out, o)
		}
	}
	return out
}

func indexAfter(s, sep string) int {
	for i := 0; i+len(sep) <= len(s); i++ {
		if s[i:i+len(sep)] == sep {
			return i + len(sep)
		}
	}
	return -1
}
