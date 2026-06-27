package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/localnarrator"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/core/signal"
	"github.com/xcreativs/gigmann/internal/seed"
)

// TestBriefQualityHarness asserts the brief invariants across several synthetic
// networks: items are worst-first and every item references a real facility (the
// deterministic-narration contract — no invented entities).
func TestBriefQualityHarness(t *testing.T) {
	asOf := time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC)
	for _, seedVal := range []int64{42, 7, 99} {
		net := seed.Generate(seedVal, asOf, seed.DefaultDays)
		svc := app.NewBriefService(signal.Default(signal.DefaultThresholds()), localnarrator.New(), 5)
		b, err := svc.Generate(context.Background(), signal.Input{
			AsOf: net.Metrics[0].Date, Facilities: net.Facilities, Metrics: net.Metrics,
			Inventory: net.Inventory, Staff: net.Staff,
		})
		require.NoError(t, err)
		require.NotEmptyf(t, b.Items, "seed %d should surface items", seedVal)
		require.NotEmpty(t, b.Prose)

		for i := 1; i < len(b.Items); i++ {
			assert.GreaterOrEqualf(t, b.Items[i-1].Severity.Rank(), b.Items[i].Severity.Rank(),
				"items must be worst-first (seed %d)", seedVal)
		}
		ids := make(map[string]bool, len(net.Facilities))
		for _, f := range net.Facilities {
			ids[f.ID] = true
		}
		for _, it := range b.Items {
			assert.Truef(t, ids[it.FacilityID], "item facility %q must be in the network (seed %d)", it.FacilityID, seedVal)
			assert.NotEmpty(t, it.Headline)
		}
	}
}

// TestBriefLeadsWithPlantedCriticalStory: the Appendix-C demo seed must lead with
// a critical item (the Tafo revenue/claims story).
func TestBriefLeadsWithPlantedCriticalStory(t *testing.T) {
	net := seed.Generate(42, time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC), seed.DefaultDays)
	svc := app.NewBriefService(signal.Default(signal.DefaultThresholds()), localnarrator.New(), 5)
	b, err := svc.Generate(context.Background(), signal.Input{
		AsOf: net.Metrics[0].Date, Facilities: net.Facilities, Metrics: net.Metrics,
		Inventory: net.Inventory, Staff: net.Staff,
	})
	require.NoError(t, err)
	require.NotEmpty(t, b.Items)
	assert.Equal(t, severity.Critical, b.Items[0].Severity, "the worst item leads")

	var hasTafo bool
	for _, it := range b.Items {
		if it.FacilityID == "tafo-maternity" {
			hasTafo = true
		}
	}
	assert.True(t, hasTafo, "the planted Tafo story surfaces")
}
