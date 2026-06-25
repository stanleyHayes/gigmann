package metric_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/core/metric"
	"github.com/xcreativs/gigmann/internal/core/money"
)

func valid() metric.FacilityMetric {
	return metric.FacilityMetric{
		FacilityID: "tafo-maternity", Date: time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC),
		Revenue: money.FromCedis(350000, 0), PatientsSeen: 3900, OccupancyRate: 0.72,
		NHISClaimsSubmitted: 100, NHISClaimsDenied: 8,
	}
}

func TestNewValid(t *testing.T) {
	m, err := metric.New(valid())
	require.NoError(t, err)
	assert.Equal(t, "tafo-maternity", m.FacilityID)
}

func TestNewInvariants(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(m *metric.FacilityMetric)
		wantErr error
	}{
		{"empty facility", func(m *metric.FacilityMetric) { m.FacilityID = " " }, metric.ErrEmptyFacilityID},
		{"zero date", func(m *metric.FacilityMetric) { m.Date = time.Time{} }, metric.ErrZeroDate},
		{"negative count", func(m *metric.FacilityMetric) { m.PatientsSeen = -1 }, metric.ErrNegativeCount},
		{"bad occupancy", func(m *metric.FacilityMetric) { m.OccupancyRate = 1.5 }, metric.ErrBadOccupancy},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := valid()
			tt.mutate(&m)
			_, err := metric.New(m)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestDenialRate(t *testing.T) {
	m := valid()
	assert.InDelta(t, 0.08, m.DenialRate(), 0.0001)
	m.NHISClaimsSubmitted = 0
	assert.Zero(t, m.DenialRate())
}
