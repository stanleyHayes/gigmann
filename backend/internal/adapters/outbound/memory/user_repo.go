package memory

import (
	"context"
	"strings"
	"sync"

	"github.com/xcreativs/gigmann/internal/ports"
)

// UserRepo is an in-memory ports.UserRepository indexed by email and id.
type UserRepo struct {
	mu      sync.RWMutex
	byEmail map[string]ports.Account
	byID    map[string]ports.Account
}

// NewUserRepo creates a repository optionally seeded with accounts.
func NewUserRepo(accounts ...ports.Account) *UserRepo {
	r := &UserRepo{byEmail: map[string]ports.Account{}, byID: map[string]ports.Account{}}
	for _, a := range accounts {
		r.put(a)
	}
	return r
}

var _ ports.UserRepository = (*UserRepo)(nil)

// FindByEmail returns the account for the given email, or ErrAccountNotFound.
func (r *UserRepo) FindByEmail(_ context.Context, email string) (ports.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.byEmail[normalizeEmail(email)]
	if !ok {
		return ports.Account{}, ports.ErrAccountNotFound
	}
	return a, nil
}

// FindByID returns the account for the given user id, or ErrAccountNotFound.
func (r *UserRepo) FindByID(_ context.Context, id string) (ports.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.byID[id]
	if !ok {
		return ports.Account{}, ports.ErrAccountNotFound
	}
	return a, nil
}

// Save upserts an account (re-indexing by email and id).
func (r *UserRepo) Save(_ context.Context, account ports.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.put(account)
	return nil
}

func (r *UserRepo) put(a ports.Account) {
	r.byEmail[normalizeEmail(a.Email)] = a
	r.byID[a.User.ID] = a
}

func normalizeEmail(e string) string { return strings.ToLower(strings.TrimSpace(e)) }
