package metric_test

import (
	"errors"
	"testing"
	"time"

	"github.com/xcreativs/gigmann/internal/core/metric"
	"github.com/xcreativs/gigmann/internal/core/money"
)

func valid() metric.FacilityMetric {
	return metric.FacilityMetric{
		FacilityID:          "tafo-maternity",
		Date:                time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC),
		Revenue:             money.FromCedis(350000, 0),
		PatientsSeen:        3900,
		OccupancyRate:       0.72,
		NHISClaimsSubmitted: 100,
		NHISClaimsDenied:    8,
	}
}

func TestNewValid(t *testing.T) {
	m, err := metric.New(valid())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if m.FacilityID != "tafo-maternity" {
		t.Errorf("facility id not set: %q", m.FacilityID)
	}
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
			if _, err := metric.New(m); !errors.Is(err, tt.wantErr) {
				t.Fatalf("want %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestDenialRate(t *testing.T) {
	m := valid()
	if got := m.DenialRate(); got != 0.08 {
		t.Errorf("want 0.08, got %v", got)
	}
	m.NHISClaimsSubmitted = 0
	if got := m.DenialRate(); got != 0 {
		t.Errorf("want 0 with no submissions, got %v", got)
	}
}
