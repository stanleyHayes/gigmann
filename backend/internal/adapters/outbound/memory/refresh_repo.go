package memory

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/ports"
)

const refreshTokenBytes = 32

type refreshRecord struct {
	principal auth.Principal
	expiresAt time.Time
}

// RefreshStore is an in-memory ports.RefreshTokenStore. Raw tokens are returned
// once and never persisted — only their SHA-256 hashes are stored, and each
// token is single-use (consumed on refresh, which rotates it).
type RefreshStore struct {
	mu     sync.Mutex
	byHash map[string]refreshRecord
}

// NewRefreshStore creates an empty in-memory refresh-token store.
func NewRefreshStore() *RefreshStore {
	return &RefreshStore{byHash: map[string]refreshRecord{}}
}

var _ ports.RefreshTokenStore = (*RefreshStore)(nil)

// Issue mints a random refresh token, stores its hash, and returns the raw token.
func (s *RefreshStore) Issue(_ context.Context, p auth.Principal, ttl time.Duration) (string, error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("memory: mint refresh token: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(buf)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byHash[hashToken(raw)] = refreshRecord{principal: p, expiresAt: time.Now().Add(ttl)}
	return raw, nil
}

// Consume validates and single-use-consumes a refresh token, returning its principal.
func (s *RefreshStore) Consume(_ context.Context, raw string) (auth.Principal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	h := hashToken(raw)
	rec, ok := s.byHash[h]
	if !ok {
		return auth.Principal{}, ports.ErrInvalidRefreshToken
	}
	delete(s.byHash, h) // single-use: rotation invalidates the presented token
	if time.Now().After(rec.expiresAt) {
		return auth.Principal{}, ports.ErrInvalidRefreshToken
	}
	return rec.principal, nil
}

// Revoke deletes a refresh token (logout); revoking an unknown token is a no-op.
func (s *RefreshStore) Revoke(_ context.Context, raw string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byHash, hashToken(raw))
	return nil
}

// RevokeUser deletes all live refresh tokens for a user. Sensitive account
// changes call this to force new sessions through the current auth policy.
func (s *RefreshStore) RevokeUser(_ context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for h, rec := range s.byHash {
		if rec.principal.UserID == userID {
			delete(s.byHash, h)
		}
	}
	return nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
