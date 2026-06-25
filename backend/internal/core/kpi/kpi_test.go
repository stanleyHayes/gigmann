package kpi_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/core/kpi"
	"github.com/xcreativs/gigmann/internal/core/metric"
	"github.com/xcreativs/gigmann/internal/core/money"
	"github.com/xcreativs/gigmann/internal/seed"
)

func find(t *testing.T, n kpi.Network, key string) kpi.KPI {
	t.Helper()
	for _, k := range n.KPIs {
		if k.Key == key {
			return k
		}
	}
	t.Fatalf("kpi %q not found", key)
	return kpi.KPI{}
}

func TestComputeExact(t *testing.T) {
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	rev := []int64{100, 100, 200, 200} // cedis
	den := []int{1, 1, 3, 3}
	ms := make([]metric.FacilityMetric, 0, 4)
	for i := range 4 {
		ms = append(ms, metric.FacilityMetric{
			FacilityID: "f1", Date: base.AddDate(0, 0, i),
			Revenue: money.FromCedis(rev[i], 0), PatientsSeen: 10, OccupancyRate: 0.5,
			NHISClaimsSubmitted: 10, NHISClaimsDenied: den[i],
		})
	}

	n := kpi.Compute(ms)
	require.Len(t, n.KPIs, 4)
	assert.Equal(t, base.AddDate(0, 0, 3), n.AsOf)

	revenue := find(t, n, "revenue") // last2 = 40000 pesewas, prior2 = 20000
	assert.InDelta(t, 40000, revenue.Current, 1e-6)
	assert.InDelta(t, 20000, revenue.Previous, 1e-6)
	assert.InDelta(t, 1.0, revenue.DeltaPct, 1e-9)
	assert.Equal(t, kpi.Up, revenue.Direction)
	assert.True(t, revenue.HigherIsBetter)
	assert.Len(t, revenue.Series, 4)

	denial := find(t, n, "denial_rate") // last2 = 6/20 = .3, prior2 = 2/20 = .1
	assert.InDelta(t, 0.3, denial.Current, 1e-9)
	assert.InDelta(t, 0.1, denial.Previous, 1e-9)
	assert.InDelta(t, 2.0, denial.DeltaPct, 1e-9)
	assert.Equal(t, kpi.Up, denial.Direction)
	assert.False(t, denial.HigherIsBetter)

	patients := find(t, n, "patients")
	assert.Equal(t, kpi.Flat, patients.Direction)
	assert.InDelta(t, 0, patients.DeltaPct, 1e-9)
}

func TestComputeEmpty(t *testing.T) {
	n := kpi.Compute(nil)
	require.Len(t, n.KPIs, 4)
	for _, k := range n.KPIs {
		assert.Empty(t, k.Series)
		assert.Equal(t, kpi.Flat, k.Direction)
	}
}

func TestComputeFromSeed(t *testing.T) {
	net := seed.Generate(7, time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC), 14)
	n := kpi.Compute(net.Metrics)
	require.Len(t, n.KPIs, 4)
	for _, k := range n.KPIs {
		assert.Len(t, k.Series, 14, k.Key)
	}
	for _, p := range find(t, n, "denial_rate").Series {
		assert.GreaterOrEqual(t, p.Value, 0.0)
		assert.LessOrEqual(t, p.Value, 1.0)
	}
}
