package app_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/localembedder"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/payer"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/ports"
)

// spyEmbedder records the texts it was asked to embed (to assert input bounds).
type spyEmbedder struct{ texts []string }

func (e *spyEmbedder) Embed(_ context.Context, texts []string, _ ports.EmbedKind) ([][]float32, error) {
	e.texts = texts
	return [][]float32{{0.1, 0.2, 0.3}}, nil
}

func (*spyEmbedder) Dimensions() int { return 3 }

func fac(t *testing.T, id, name, region, town, typ, mgr string) facility.Facility {
	t.Helper()
	mix, err := payer.New(70, 25, 5)
	require.NoError(t, err)
	f, err := facility.New(facility.Params{
		ID: id, Name: name, Region: facility.Region(region), Town: town, Type: typ,
		Lifecycle: facility.LifecycleActive, Health: severity.Good, ManagerName: mgr, PayerMix: mix,
	})
	require.NoError(t, err)
	return f
}

func TestFacilitySearchResolvesByName(t *testing.T) {
	ctx := context.Background()
	facilities := []facility.Facility{
		fac(t, "kasoa", "Kasoa Polyclinic", "Central", "Kasoa", "High-volume OPD", "Ama Owusu"),
		fac(t, "tamale-north", "Tamale North Clinic", "Northern", "Tamale", "General", "Fuseini Abdulai"),
		fac(t, "nima", "Nima Urban Health Centre", "Greater Accra", "Nima", "Urban", "Mohammed Iddrisu"),
	}
	embedder := localembedder.New()
	repo := memory.NewFacilityEmbeddingRepo()

	seeded, err := app.SeedFacilityEmbeddings(ctx, embedder, repo, facilities)
	require.NoError(t, err)
	assert.True(t, seeded)
	// Idempotent: a second run is a no-op.
	again, err := app.SeedFacilityEmbeddings(ctx, embedder, repo, facilities)
	require.NoError(t, err)
	assert.False(t, again)

	svc := app.NewFacilitySearchService(embedder, repo, facilities)
	got, err := svc.Resolve(ctx, "how is the Kasoa polyclinic doing", 3)
	require.NoError(t, err)
	require.NotEmpty(t, got)
	assert.Equal(t, "kasoa", got[0].FacilityID, "top match resolves to Kasoa")
	assert.Equal(t, "Kasoa Polyclinic", got[0].Name)
	assert.Greater(t, got[0].Score, 0.0)

	empty, err := svc.Resolve(ctx, "   ", 3)
	require.NoError(t, err)
	assert.Empty(t, empty)
}

func TestFacilitySearchBoundsQueryLength(t *testing.T) {
	spy := &spyEmbedder{}
	svc := app.NewFacilitySearchService(spy, memory.NewFacilityEmbeddingRepo(), nil)

	_, err := svc.Resolve(context.Background(), strings.Repeat("a", 300), 3)
	require.NoError(t, err)
	require.Len(t, spy.texts, 1)
	assert.Len(t, []rune(spy.texts[0]), 256, "over-long query is truncated to the cap before embedding")
}
