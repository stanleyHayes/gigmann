package app_test

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/ports"
)

func prefAccount(t *testing.T) ports.Account {
	t.Helper()
	u, err := user.New(user.User{ID: "u1", Name: "Sammy", Role: user.RoleExecutive})
	require.NoError(t, err)
	return ports.Account{User: u, Email: "ceo@gigmann.health", PasswordHash: "h"}
}

func TestPreferencesGetDefaultsEmpty(t *testing.T) {
	repo := memory.NewUserRepo(prefAccount(t))
	svc := app.NewPreferencesService(repo)
	got, err := svc.Get(context.Background(), "u1")
	require.NoError(t, err)
	assert.Empty(t, got.WatchedMetrics)
	assert.Empty(t, got.Thresholds)
}

func TestPreferencesUpdateRoundTrips(t *testing.T) {
	repo := memory.NewUserRepo(prefAccount(t))
	svc := app.NewPreferencesService(repo)
	updated, err := svc.Update(context.Background(), "u1", user.Preferences{
		WatchedMetrics: []string{"revenue", "denial_rate"},
		Thresholds:     map[string]float64{"denial_rate": 0.15},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"revenue", "denial_rate"}, updated.WatchedMetrics)
	assert.InDelta(t, 0.15, updated.Thresholds["denial_rate"], 1e-9)

	// Persisted: a fresh Get returns the same.
	got, err := svc.Get(context.Background(), "u1")
	require.NoError(t, err)
	assert.Equal(t, updated, got)
}

func TestPreferencesUpdateSanitizes(t *testing.T) {
	repo := memory.NewUserRepo(prefAccount(t))
	svc := app.NewPreferencesService(repo)
	updated, err := svc.Update(context.Background(), "u1", user.Preferences{
		WatchedMetrics: []string{" revenue ", "revenue", "", "occupancy"},
		Thresholds:     map[string]float64{"x": math.Inf(1), "y": 0.2, " ": 1},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"revenue", "occupancy"}, updated.WatchedMetrics, "trimmed + de-duped + empties dropped")
	assert.NotContains(t, updated.Thresholds, "x", "non-finite dropped")
	assert.NotContains(t, updated.Thresholds, " ", "blank key dropped")
	assert.InDelta(t, 0.2, updated.Thresholds["y"], 1e-9)
}

func TestPreferencesUnknownUser(t *testing.T) {
	repo := memory.NewUserRepo()
	svc := app.NewPreferencesService(repo)
	_, err := svc.Get(context.Background(), "ghost")
	require.Error(t, err)
}
