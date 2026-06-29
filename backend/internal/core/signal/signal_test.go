package signal_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/inventory"
	"github.com/xcreativs/gigmann/internal/core/metric"
	"github.com/xcreativs/gigmann/internal/core/money"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/core/signal"
	"github.com/xcreativs/gigmann/internal/core/staff"
)

var base = time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC) // a Monday

// mk builds a valid metric for the test facility on base+day.
func mk(t *testing.T, day int, revPesewas int64, patients, submitted, denied int, unbilled int64) metric.FacilityMetric {
	t.Helper()
	m, err := metric.New(metric.FacilityMetric{
		FacilityID: "f", Date: base.AddDate(0, 0, day),
		Revenue: money.FromPesewas(revPesewas), PatientsSeen: patients,
		NHISClaimsSubmitted: submitted, NHISClaimsDenied: denied,
		UnbilledAmount: money.FromPesewas(unbilled), OccupancyRate: 0.5,
	})
	require.NoError(t, err)
	return m
}

// series builds 14 days: 7 "prev" then 7 "recent", each day identical within its week.
func series(t *testing.T, prevRev, recentRev int64, prevPat, recentPat, prevSub, recentSub, denied int, unbilled int64) []metric.FacilityMetric {
	t.Helper()
	ms := make([]metric.FacilityMetric, 0, 14)
	for d := range 14 {
		if d < 7 {
			ms = append(ms, mk(t, d, prevRev, prevPat, prevSub, 0, unbilled))
		} else {
			ms = append(ms, mk(t, d, recentRev, recentPat, recentSub, denied, unbilled))
		}
	}
	return ms
}

type stubDetector struct{ sigs []signal.Signal }

func (s stubDetector) Name() string                        { return "stub" }
func (s stubDetector) Detect(signal.Input) []signal.Signal { return s.sigs }

func TestEngineRunRanksWorstFirst(t *testing.T) {
	eng := signal.NewEngine(stubDetector{sigs: []signal.Signal{
		{Type: "b", FacilityID: "f2", Severity: severity.Watch, Magnitude: 0.2},
		{Type: "a", FacilityID: "f1", Severity: severity.Critical, Magnitude: 0.5},
		{Type: "a", FacilityID: "f1", Severity: severity.Critical, Magnitude: 0.9},
	}})
	got := eng.Run(signal.Input{})
	require.Len(t, got, 3)
	assert.Equal(t, severity.Critical, got[0].Severity)
	assert.InDelta(t, 0.9, got[0].Magnitude, 0.001) // highest magnitude critical first
	assert.Equal(t, severity.Watch, got[2].Severity)
}

func TestDefaultThresholds(t *testing.T) {
	th := signal.DefaultThresholds()
	assert.Greater(t, th.RevenueDropCritical, th.RevenueDropWatch)
	assert.Greater(t, th.UnbilledCritical, th.UnbilledWatch)
}

// TestRankDeterministicTiebreak covers the lower tiebreakers: with equal
// severity and magnitude, order is FacilityID asc, then Type asc — so the brief
// is stable and reproducible run-to-run.
func TestRankDeterministicTiebreak(t *testing.T) {
	eng := signal.NewEngine(stubDetector{sigs: []signal.Signal{
		{Type: "z", FacilityID: "f2", Severity: severity.Watch, Magnitude: 0.5},
		{Type: "b", FacilityID: "f1", Severity: severity.Watch, Magnitude: 0.5},
		{Type: "a", FacilityID: "f1", Severity: severity.Watch, Magnitude: 0.5},
	}})
	got := eng.Run(signal.Input{})
	require.Len(t, got, 3)
	assert.Equal(t, "f1", got[0].FacilityID)
	assert.Equal(t, "a", got[0].Type, "within a facility, Type breaks the tie ascending")
	assert.Equal(t, "f1", got[1].FacilityID)
	assert.Equal(t, "b", got[1].Type)
	assert.Equal(t, "f2", got[2].FacilityID, "FacilityID breaks the tie ascending")
}

func TestDetectorNames(t *testing.T) {
	th := signal.DefaultThresholds()
	assert.Equal(t, "leakage", signal.NewLeakageDetector(th).Name())
	assert.Equal(t, "claims", signal.NewClaimsDetector(th).Name())
	assert.Equal(t, "stockout", signal.NewStockOutDetector().Name())
	assert.Equal(t, "staff", signal.NewStaffDetector(th).Name())
	assert.Equal(t, "trend", signal.NewTrendDetector(th).Name())
}

func TestTrendDetector(t *testing.T) {
	th := signal.DefaultThresholds()
	d := signal.NewTrendDetector(th)

	// −30% → critical
	crit := d.Detect(signal.Input{Metrics: series(t, 1000, 700, 100, 100, 60, 60, 0, 0)})
	require.Len(t, crit, 1)
	assert.Equal(t, severity.Critical, crit[0].Severity)
	assert.Equal(t, "revenue_drop", crit[0].Type)

	// −12% → watch
	watch := d.Detect(signal.Input{Metrics: series(t, 1000, 880, 100, 100, 60, 60, 0, 0)})
	require.Len(t, watch, 1)
	assert.Equal(t, severity.Watch, watch[0].Severity)

	// −2% → none
	none := d.Detect(signal.Input{Metrics: series(t, 1000, 980, 100, 100, 60, 60, 0, 0)})
	assert.Empty(t, none)

	// zero previous revenue → skipped
	zero := d.Detect(signal.Input{Metrics: series(t, 0, 0, 100, 100, 60, 60, 0, 0)})
	assert.Empty(t, zero)

	// no metrics → no signals
	assert.Empty(t, d.Detect(signal.Input{}))
}

func TestClaimsDetector(t *testing.T) {
	th := signal.DefaultThresholds()
	d := signal.NewClaimsDetector(th)

	// denial rate 20/100 = 20% → critical denial_spike
	den := d.Detect(signal.Input{Metrics: series(t, 1000, 1000, 100, 100, 100, 100, 20, 0)})
	assert.True(t, hasType(den, "denial_spike"))
	for _, s := range den {
		if s.Type == "denial_spike" {
			assert.Equal(t, severity.Critical, s.Severity)
		}
	}

	// submission gap: patients flat, submissions collapse 60 → 20 (−66%) → critical
	gap := d.Detect(signal.Input{Metrics: series(t, 1000, 1000, 100, 100, 60, 20, 0, 0)})
	assert.True(t, hasType(gap, "submission_gap"))

	// healthy: no denial, no gap
	ok := d.Detect(signal.Input{Metrics: series(t, 1000, 1000, 100, 100, 60, 60, 1, 0)})
	assert.False(t, hasType(ok, "denial_spike"))
	assert.False(t, hasType(ok, "submission_gap"))
}

func TestClaimsDetectorGuards(t *testing.T) {
	d := signal.NewClaimsDetector(signal.DefaultThresholds())

	// 14% denial (between Watch 0.10 and Critical 0.18) → Watch-tier denial_spike.
	watch := d.Detect(signal.Input{Metrics: series(t, 1000, 1000, 100, 100, 100, 100, 14, 0)})
	require.True(t, hasType(watch, "denial_spike"))
	for _, s := range watch {
		if s.Type == "denial_spike" {
			assert.Equal(t, severity.Watch, s.Severity)
		}
	}

	// Zero submissions this week → no denial_spike (can't compute a rate), but the
	// collapse 60→0 is still a submission_gap.
	zeroSub := d.Detect(signal.Input{Metrics: series(t, 1000, 1000, 100, 100, 60, 0, 0, 0)})
	assert.False(t, hasType(zeroSub, "denial_spike"))
	assert.True(t, hasType(zeroSub, "submission_gap"))

	// No prior-week activity (patients & submissions both 0) → no submission_gap.
	noPrev := d.Detect(signal.Input{Metrics: series(t, 1000, 1000, 0, 100, 0, 20, 0, 0)})
	assert.False(t, hasType(noPrev, "submission_gap"))

	// A single day is too little to split into two windows → no signals.
	one := d.Detect(signal.Input{Metrics: []metric.FacilityMetric{mk(t, 0, 1000, 100, 60, 0, 0)}})
	assert.Empty(t, one)
}

func TestLeakageDetector(t *testing.T) {
	d := signal.NewLeakageDetector(signal.DefaultThresholds())
	// 14 days * 500_000 = 7_000_000 pesewas (GH₵70k) ≥ critical 6_000_000
	crit := d.Detect(signal.Input{Metrics: series(t, 1000, 1000, 100, 100, 60, 60, 0, 500_000)})
	require.Len(t, crit, 1)
	assert.Equal(t, severity.Critical, crit[0].Severity)
	// Magnitude is normalised to the critical threshold (≈1.17 here), not raw pesewas,
	// so it ranks comparably against ratio-based signals within a severity.
	assert.InDelta(t, 7_000_000.0/6_000_000.0, crit[0].Magnitude, 1e-6)

	// 14 * 250_000 = 3_500_000 (GH₵35k) → watch
	watch := d.Detect(signal.Input{Metrics: series(t, 1000, 1000, 100, 100, 60, 60, 0, 250_000)})
	require.Len(t, watch, 1)
	assert.Equal(t, severity.Watch, watch[0].Severity)

	// below threshold
	assert.Empty(t, d.Detect(signal.Input{Metrics: series(t, 1000, 1000, 100, 100, 60, 60, 0, 1000)}))
}

func TestStockOutDetector(t *testing.T) {
	d := signal.NewStockOutDetector()
	mkItem := func(stock int, burn float64, lead int) inventory.Item {
		it, err := inventory.New(inventory.Item{ID: "i", FacilityID: "f", Name: "RDT", StockLevel: stock, DailyBurn: burn, LeadTimeDays: lead})
		require.NoError(t, err)
		return it
	}
	// 50/10 = 5 days vs 7 lead → watch (5 >= 3.5)
	watch := d.Detect(signal.Input{Inventory: []inventory.Item{mkItem(50, 10, 7)}})
	require.Len(t, watch, 1)
	assert.Equal(t, severity.Watch, watch[0].Severity)
	// 10/10 = 1 day vs 7 lead → critical (1 < 3.5)
	crit := d.Detect(signal.Input{Inventory: []inventory.Item{mkItem(10, 10, 7)}})
	require.Len(t, crit, 1)
	assert.Equal(t, severity.Critical, crit[0].Severity)
	// 1 day of stock vs 7-day lead time → normalised intensity (7-1)/7, not raw days.
	assert.InDelta(t, 6.0/7.0, crit[0].Magnitude, 1e-6)
	// well stocked → none
	assert.Empty(t, d.Detect(signal.Input{Inventory: []inventory.Item{mkItem(1000, 10, 7)}}))
}

func TestStaffDetector(t *testing.T) {
	d := signal.NewStaffDetector(signal.DefaultThresholds())
	asOf := base
	mkStaff := func(expiry time.Time, risk float64) staff.Member {
		m, err := staff.New(staff.Member{ID: "s", FacilityID: "f", Name: "X", Role: "Nurse", LicenceExpiry: expiry, AttritionRisk: risk})
		require.NoError(t, err)
		return m
	}
	// licence expiring in 10 days + high attrition → two signals
	both := d.Detect(signal.Input{AsOf: asOf, Staff: []staff.Member{mkStaff(asOf.AddDate(0, 0, 10), 0.8)}})
	assert.True(t, hasType(both, "licence_expiry"))
	assert.True(t, hasType(both, "attrition_risk"))
	// healthy: licence far out, low risk → none
	none := d.Detect(signal.Input{AsOf: asOf, Staff: []staff.Member{mkStaff(asOf.AddDate(1, 0, 0), 0.1)}})
	assert.Empty(t, none)

	// Magnitude reflects urgency: a sooner expiry must outrank a later one
	// (not a flat constant) so the brief prioritises it correctly.
	licMag := func(sigs []signal.Signal) float64 {
		for _, s := range sigs {
			if s.Type == "licence_expiry" {
				return s.Magnitude
			}
		}
		t.Fatal("expected a licence_expiry signal")
		return 0
	}
	soon := d.Detect(signal.Input{AsOf: asOf, Staff: []staff.Member{mkStaff(asOf.AddDate(0, 0, 2), 0)}})
	later := d.Detect(signal.Input{AsOf: asOf, Staff: []staff.Member{mkStaff(asOf.AddDate(0, 0, 9), 0)}})
	assert.Greater(t, licMag(soon), licMag(later), "a sooner licence expiry ranks above a later one")
}

func TestNetworkPulse(t *testing.T) {
	facs := []facility.Facility{}
	good := signal.NetworkPulse(facs, nil)
	assert.Equal(t, severity.Good, good.Severity)
	assert.InDelta(t, 1.0, good.Score, 0.001)

	watch := signal.NetworkPulse(facs, []signal.Signal{{Severity: severity.Watch}, {Severity: severity.Good}})
	assert.Equal(t, severity.Watch, watch.Severity)
	assert.Equal(t, 1, watch.WatchCount)

	crit := signal.NetworkPulse(facs, []signal.Signal{{Severity: severity.Critical}, {Severity: severity.Watch}})
	assert.Equal(t, severity.Critical, crit.Severity)
	assert.Equal(t, 1, crit.CriticalCount)
	assert.Less(t, crit.Score, 1.0)
	assert.Contains(t, crit.Headline, "critical")
}

func hasType(sigs []signal.Signal, typ string) bool {
	for _, s := range sigs {
		if s.Type == typ {
			return true
		}
	}
	return false
}
