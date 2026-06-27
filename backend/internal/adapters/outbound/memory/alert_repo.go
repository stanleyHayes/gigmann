package memory

import (
	"context"
	"sync"

	"github.com/xcreativs/gigmann/internal/core/alert"
	"github.com/xcreativs/gigmann/internal/ports"
)

// AlertRepo is an in-memory ports.AlertRepository preserving seed order.
type AlertRepo struct {
	mu    sync.RWMutex
	items []alert.Alert
}

// NewAlertRepo creates a repository optionally seeded with alerts.
func NewAlertRepo(seed ...alert.Alert) *AlertRepo {
	return &AlertRepo{items: append([]alert.Alert{}, seed...)}
}

var _ ports.AlertRepository = (*AlertRepo)(nil)

// List returns a copy of all alerts.
func (r *AlertRepo) List(_ context.Context) ([]alert.Alert, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]alert.Alert, len(r.items))
	copy(out, r.items)
	return out, nil
}

// Get returns the alert with the given id, or ErrAlertNotFound.
func (r *AlertRepo) Get(_ context.Context, id string) (alert.Alert, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.items {
		if a.ID == id {
			return a, nil
		}
	}
	return alert.Alert{}, ports.ErrAlertNotFound
}

// Save updates an existing alert in place (or appends a new one).
func (r *AlertRepo) Save(_ context.Context, a alert.Alert) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == a.ID {
			r.items[i] = a
			return nil
		}
	}
	r.items = append(r.items, a)
	return nil
}
