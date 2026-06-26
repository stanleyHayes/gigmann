package memory

import (
	"context"
	"sync"

	"github.com/xcreativs/gigmann/internal/core/approval"
	"github.com/xcreativs/gigmann/internal/ports"
)

// ApprovalRepo is an in-memory ports.ApprovalRepository preserving seed order.
type ApprovalRepo struct {
	mu    sync.RWMutex
	items []approval.Approval
}

// NewApprovalRepo creates a repository optionally seeded with approvals.
func NewApprovalRepo(seed ...approval.Approval) *ApprovalRepo {
	return &ApprovalRepo{items: append([]approval.Approval{}, seed...)}
}

var _ ports.ApprovalRepository = (*ApprovalRepo)(nil)

// List returns a copy of all approvals.
func (r *ApprovalRepo) List(_ context.Context) ([]approval.Approval, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]approval.Approval, len(r.items))
	copy(out, r.items)
	return out, nil
}

// Get returns the approval with the given id, or ErrApprovalNotFound.
func (r *ApprovalRepo) Get(_ context.Context, id string) (approval.Approval, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.items {
		if a.ID == id {
			return a, nil
		}
	}
	return approval.Approval{}, ports.ErrApprovalNotFound
}

// Save updates an existing approval in place (or appends a new one).
func (r *ApprovalRepo) Save(_ context.Context, a approval.Approval) error {
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
