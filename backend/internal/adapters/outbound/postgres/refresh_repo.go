package postgres

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/postgres/sqlcgen"
	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/ports"
)

// RefreshRepo is a PostgreSQL implementation of ports.RefreshTokenStore. Only the
// SHA-256 hash of a token is persisted, and tokens are single-use: Consume deletes
// the row (RETURNING the principal) so a rotated or replayed token cannot be reused.
type RefreshRepo struct {
	q *sqlcgen.Queries
}

var _ ports.RefreshTokenStore = (*RefreshRepo)(nil)

// NewRefreshRepo builds a RefreshRepo over a pgx pool (or any sqlcgen.DBTX).
func NewRefreshRepo(db sqlcgen.DBTX) *RefreshRepo {
	return &RefreshRepo{q: sqlcgen.New(db)}
}

// Issue mints a random refresh token, stores its hash with the principal snapshot,
// and returns the raw token (shown to the client once).
func (r *RefreshRepo) Issue(ctx context.Context, p auth.Principal, ttl time.Duration) (string, error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("postgres: mint refresh token: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(buf)
	if err := r.q.InsertRefreshToken(ctx, sqlcgen.InsertRefreshTokenParams{
		TokenHash:  hashToken(raw),
		UserID:     p.UserID,
		Name:       p.Name,
		Role:       string(p.Role),
		FacilityID: p.FacilityID,
		ExpiresAt:  tsRequired(time.Now().Add(ttl)),
	}); err != nil {
		return "", fmt.Errorf("postgres: insert refresh token: %w", err)
	}
	return raw, nil
}

// Consume validates and single-use-consumes a refresh token, returning its principal.
func (r *RefreshRepo) Consume(ctx context.Context, raw string) (auth.Principal, error) {
	row, err := r.q.ConsumeRefreshToken(ctx, hashToken(raw))
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.Principal{}, ports.ErrInvalidRefreshToken
	}
	if err != nil {
		return auth.Principal{}, fmt.Errorf("postgres: consume refresh token: %w", err)
	}
	// The row is deleted regardless; an expired token is therefore also cleaned up.
	if time.Now().After(timeFromTS(row.ExpiresAt)) {
		return auth.Principal{}, ports.ErrInvalidRefreshToken
	}
	return auth.Principal{
		UserID:     row.UserID,
		Name:       row.Name,
		Role:       user.Role(row.Role),
		FacilityID: row.FacilityID,
	}, nil
}

// Revoke deletes a refresh token (logout); revoking an unknown token is a no-op.
func (r *RefreshRepo) Revoke(ctx context.Context, raw string) error {
	if err := r.q.DeleteRefreshToken(ctx, hashToken(raw)); err != nil {
		return fmt.Errorf("postgres: revoke refresh token: %w", err)
	}
	return nil
}

// RevokeUser deletes all refresh tokens for a user after sensitive account changes.
func (r *RefreshRepo) RevokeUser(ctx context.Context, userID string) error {
	if err := r.q.DeleteRefreshTokensForUser(ctx, userID); err != nil {
		return fmt.Errorf("postgres: revoke user refresh tokens: %w", err)
	}
	return nil
}
