package memory_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/payer"
	"github.com/xcreativs/gigmann/internal/core/severity"
)

func mustFacility(t *testing.T) facility.Facility {
	t.Helper()
	mix, err := payer.New(70, 25, 5)
	require.NoError(t, err)
	f, err := facility.New(facility.Params{
		ID: "f1", Name: "Kasoa", Region: "Central", Town: "Kasoa",
		Type: "OPD", Beds: 40, Lifecycle: facility.LifecycleActive, Health: severity.Good,
		ManagerName: "Ama Owusu", PayerMix: mix,
	})
	require.NoError(t, err)
	return f
}

func TestFacilityRepoList(t *testing.T) {
	repo := memory.NewFacilityRepo(mustFacility(t))

	got, err := repo.List(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)

	// Returned slice must be a copy: mutating it must not affect the repo.
	got[0].Name = "mutated"
	again, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.NotEqual(t, "mutated", again[0].Name, "repo leaked its internal slice")
}

func TestFacilityRepoEmpty(t *testing.T) {
	repo := memory.NewFacilityRepo()

	got, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, got)
}
