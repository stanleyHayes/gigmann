// Package metric holds the time-series FacilityMetric entity (spec §7).
package metric

import (
	"errors"
	"strings"
	"time"

	"github.com/xcreativs/gigmann/internal/core/money"
)

// FacilityMetric is one facility's metrics for a single day/period.
type FacilityMetric struct {
	FacilityID          string
	Date                time.Time
	Revenue             money.Cedis
	CashRevenue         money.Cedis
	MoMoRevenue         money.Cedis
	PatientsSeen        int
	Admissions          int
	OccupancyRate       float64 // 0..1
	AvgWaitMinutes      int
	NHISClaimsSubmitted int
	NHISClaimsPaid      int
	NHISClaimsDenied    int
	NHISOutstanding     money.Cedis
	UnbilledAmount      money.Cedis
}

// Validation errors.
var (
	ErrEmptyFacilityID = errors.New("metric: facility_id is required")
	ErrZeroDate        = errors.New("metric: date is required")
	ErrNegativeCount   = errors.New("metric: counts must be >= 0")
	ErrBadOccupancy    = errors.New("metric: occupancy_rate must be within 0..1")
)

// New validates and returns a FacilityMetric.
func New(m FacilityMetric) (FacilityMetric, error) {
	m.FacilityID = strings.TrimSpace(m.FacilityID)
	switch {
	case m.FacilityID == "":
		return FacilityMetric{}, ErrEmptyFacilityID
	case m.Date.IsZero():
		return FacilityMetric{}, ErrZeroDate
	case m.PatientsSeen < 0, m.Admissions < 0, m.AvgWaitMinutes < 0,
		m.NHISClaimsSubmitted < 0, m.NHISClaimsPaid < 0, m.NHISClaimsDenied < 0:
		return FacilityMetric{}, ErrNegativeCount
	case m.OccupancyRate < 0 || m.OccupancyRate > 1:
		return FacilityMetric{}, ErrBadOccupancy
	}
	return m, nil
}

// DenialRate is denied / submitted claims (0 when none submitted).
func (m FacilityMetric) DenialRate() float64 {
	if m.NHISClaimsSubmitted == 0 {
		return 0
	}
	return float64(m.NHISClaimsDenied) / float64(m.NHISClaimsSubmitted)
}
