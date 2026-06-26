package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/ports"
)

func principal() auth.Principal {
	return auth.Principal{UserID: "u1", Name: "Sammy", Role: user.RoleExecutive}
}

func TestRefreshIssueConsumeRoundTrip(t *testing.T) {
	store := memory.NewRefreshStore()
	raw, err := store.Issue(context.Background(), principal(), time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, raw)

	got, err := store.Consume(context.Background(), raw)
	require.NoError(t, err)
	assert.Equal(t, "u1", got.UserID)
}

func TestRefreshIsSingleUse(t *testing.T) {
	store := memory.NewRefreshStore()
	raw, err := store.Issue(context.Background(), principal(), time.Hour)
	require.NoError(t, err)

	_, err = store.Consume(context.Background(), raw)
	require.NoError(t, err)
	_, err = store.Consume(context.Background(), raw) // already rotated
	assert.ErrorIs(t, err, ports.ErrInvalidRefreshToken)
}

func TestRefreshRejectsExpired(t *testing.T) {
	store := memory.NewRefreshStore()
	raw, err := store.Issue(context.Background(), principal(), -time.Minute)
	require.NoError(t, err)
	_, err = store.Consume(context.Background(), raw)
	assert.ErrorIs(t, err, ports.ErrInvalidRefreshToken)
}

func TestRefreshRevoke(t *testing.T) {
	store := memory.NewRefreshStore()
	raw, err := store.Issue(context.Background(), principal(), time.Hour)
	require.NoError(t, err)
	require.NoError(t, store.Revoke(context.Background(), raw))
	_, err = store.Consume(context.Background(), raw)
	assert.ErrorIs(t, err, ports.ErrInvalidRefreshToken)
}

func TestRefreshConsumeUnknown(t *testing.T) {
	store := memory.NewRefreshStore()
	_, err := store.Consume(context.Background(), "never-issued")
	assert.ErrorIs(t, err, ports.ErrInvalidRefreshToken)
}
