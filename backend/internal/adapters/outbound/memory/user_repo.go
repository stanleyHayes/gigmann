package memory

import (
	"context"
	"strings"
	"sync"

	"github.com/xcreativs/gigmann/internal/ports"
)

// UserRepo is an in-memory ports.UserRepository keyed by lower-cased email.
type UserRepo struct {
	mu      sync.RWMutex
	byEmail map[string]ports.Account
}

// NewUserRepo creates a repository optionally seeded with accounts.
func NewUserRepo(accounts ...ports.Account) *UserRepo {
	m := make(map[string]ports.Account, len(accounts))
	for _, a := range accounts {
		m[normalizeEmail(a.Email)] = a
	}
	return &UserRepo{byEmail: m}
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

func normalizeEmail(e string) string { return strings.ToLower(strings.TrimSpace(e)) }
