package memory

import (
	"context"
	"sync"

	"github.com/xcreativs/gigmann/internal/ports"
)

// PushRepo is an in-memory ports.PushSubscriptionStore. Subscriptions are keyed
// by user and deduped by endpoint, so re-subscribing the same browser is
// idempotent. It is safe for concurrent use.
type PushRepo struct {
	mu     sync.RWMutex
	byUser map[string]map[string]ports.PushSubscription // userID -> endpoint -> sub
}

// NewPushRepo creates an empty in-memory push-subscription store.
func NewPushRepo() *PushRepo {
	return &PushRepo{byUser: map[string]map[string]ports.PushSubscription{}}
}

var _ ports.PushSubscriptionStore = (*PushRepo)(nil)

// Save stores (or replaces) a subscription for the user, deduped by endpoint.
func (r *PushRepo) Save(_ context.Context, userID string, sub ports.PushSubscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	subs := r.byUser[userID]
	if subs == nil {
		subs = map[string]ports.PushSubscription{}
		r.byUser[userID] = subs
	}
	subs[sub.Endpoint] = sub
	return nil
}

// Delete removes the subscription with the given endpoint for the user (no-op if absent).
func (r *PushRepo) Delete(_ context.Context, userID, endpoint string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if subs := r.byUser[userID]; subs != nil {
		delete(subs, endpoint)
		if len(subs) == 0 {
			delete(r.byUser, userID)
		}
	}
	return nil
}

// ListByUser returns the user's current subscriptions.
func (r *PushRepo) ListByUser(_ context.Context, userID string) ([]ports.PushSubscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return collect(r.byUser[userID]), nil
}

// All returns every subscription grouped by user (used by the critical-alert sweep).
func (r *PushRepo) All(_ context.Context) (map[string][]ports.PushSubscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string][]ports.PushSubscription, len(r.byUser))
	for userID, subs := range r.byUser {
		out[userID] = collect(subs)
	}
	return out, nil
}

func collect(subs map[string]ports.PushSubscription) []ports.PushSubscription {
	out := make([]ports.PushSubscription, 0, len(subs))
	for _, s := range subs {
		out = append(out, s)
	}
	return out
}
