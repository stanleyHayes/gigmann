package memory

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/xcreativs/gigmann/internal/ports"
)

const passwordResetTokenBytes = 32

type passwordResetRecord struct {
	userID    string
	expiresAt time.Time
}

// PasswordResetStore is an in-memory ports.PasswordResetTokenStore. It stores
// only token hashes and consumes tokens in one step.
type PasswordResetStore struct {
	mu     sync.Mutex
	byHash map[string]passwordResetRecord
}

var _ ports.PasswordResetTokenStore = (*PasswordResetStore)(nil)

// NewPasswordResetStore creates an empty in-memory password reset token store.
func NewPasswordResetStore() *PasswordResetStore {
	return &PasswordResetStore{byHash: map[string]passwordResetRecord{}}
}

// Issue mints a short-lived reset token for a user and returns the raw token once.
func (s *PasswordResetStore) Issue(_ context.Context, userID string, ttl time.Duration) (string, error) {
	buf := make([]byte, passwordResetTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("memory: mint password reset token: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(buf)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byHash[hashToken(raw)] = passwordResetRecord{userID: userID, expiresAt: time.Now().Add(ttl)}
	return raw, nil
}

// Consume validates and single-use-consumes a reset token, returning its user id.
func (s *PasswordResetStore) Consume(_ context.Context, raw string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	h := hashToken(raw)
	rec, ok := s.byHash[h]
	if !ok {
		return "", ports.ErrInvalidPasswordResetToken
	}
	delete(s.byHash, h)
	if time.Now().After(rec.expiresAt) {
		return "", ports.ErrInvalidPasswordResetToken
	}
	return rec.userID, nil
}
