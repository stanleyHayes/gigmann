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
	"github.com/xcreativs/gigmann/internal/ports"
)

// PasswordResetRepo is a PostgreSQL implementation of ports.PasswordResetTokenStore.
type PasswordResetRepo struct {
	q *sqlcgen.Queries
}

var _ ports.PasswordResetTokenStore = (*PasswordResetRepo)(nil)

// NewPasswordResetRepo builds a PasswordResetRepo over a pgx pool (or any sqlcgen.DBTX).
func NewPasswordResetRepo(db sqlcgen.DBTX) *PasswordResetRepo {
	return &PasswordResetRepo{q: sqlcgen.New(db)}
}

// Issue mints a short-lived reset token for a user and stores only its hash.
func (r *PasswordResetRepo) Issue(ctx context.Context, userID string, ttl time.Duration) (string, error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("postgres: mint password reset token: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(buf)
	if err := r.q.InsertPasswordResetToken(ctx, sqlcgen.InsertPasswordResetTokenParams{
		TokenHash: hashToken(raw),
		UserID:    userID,
		ExpiresAt: tsRequired(time.Now().Add(ttl)),
	}); err != nil {
		return "", fmt.Errorf("postgres: insert password reset token: %w", err)
	}
	return raw, nil
}

// Consume validates and single-use-consumes a reset token, returning its user id.
func (r *PasswordResetRepo) Consume(ctx context.Context, raw string) (string, error) {
	row, err := r.q.ConsumePasswordResetToken(ctx, hashToken(raw))
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ports.ErrInvalidPasswordResetToken
	}
	if err != nil {
		return "", fmt.Errorf("postgres: consume password reset token: %w", err)
	}
	if time.Now().After(timeFromTS(row.ExpiresAt)) {
		return "", ports.ErrInvalidPasswordResetToken
	}
	return row.UserID, nil
}
