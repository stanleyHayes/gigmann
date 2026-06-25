// Package signal is the deterministic signal engine (spec §6.3). Detectors turn
// computed read models into ranked, human-readable Signals. All numbers,
// thresholds, and deltas are computed here in code — never by the AI.
package signal

import (
	"math"
	"sort"
	"time"

	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/inventory"
	"github.com/xcreativs/gigmann/internal/core/metric"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/core/staff"
)

// Signal is one flagged, ranked finding produced by a detector.
type Signal struct {
	Type              string
	FacilityID        string
	Severity          severity.Severity
	Magnitude         float64 // normalised impact, used for ranking within a severity
	Headline          string
	SupportingFigures map[string]any
}

// Input bundles the read models a detector may inspect.
type Input struct {
	AsOf       time.Time
	Facilities []facility.Facility
	Metrics    []metric.FacilityMetric
	Inventory  []inventory.Item
	Staff      []staff.Member
}

// Detector flags signals from the input. Implementations are pure.
type Detector interface {
	Name() string
	Detect(in Input) []Signal
}

// Thresholds externalises the detection cut-offs.
type Thresholds struct {
	RevenueDropWatch    float64
	RevenueDropCritical float64
	DenialRateWatch     float64
	DenialRateCritical  float64
	SubmissionGapWatch  float64
	UnbilledWatch       int64 // pesewas
	UnbilledCritical    int64 // pesewas
	LicenceWindowDays   int
	AttritionWatch      float64
}

// DefaultThresholds returns sensible defaults for the synthetic network.
func DefaultThresholds() Thresholds {
	return Thresholds{
		RevenueDropWatch:    0.10,
		RevenueDropCritical: 0.20,
		DenialRateWatch:     0.10,
		DenialRateCritical:  0.18,
		SubmissionGapWatch:  0.30,
		UnbilledWatch:       3_000_000, // GH₵ 30,000
		UnbilledCritical:    6_000_000, // GH₵ 60,000
		LicenceWindowDays:   30,
		AttritionWatch:      0.6,
	}
}

// Engine runs a set of detectors and returns their ranked signals.
type Engine struct {
	detectors []Detector
}

// NewEngine builds an Engine from the given detectors.
func NewEngine(detectors ...Detector) *Engine {
	return &Engine{detectors: detectors}
}

// Default builds an Engine with all standard detectors at the given thresholds.
func Default(th Thresholds) *Engine {
	return NewEngine(
		NewTrendDetector(th),
		NewClaimsDetector(th),
		NewLeakageDetector(th),
		NewStockOutDetector(),
		NewStaffDetector(th),
	)
}

// Run executes every detector and returns the signals ranked worst-first.
func (e *Engine) Run(in Input) []Signal {
	out := make([]Signal, 0, len(e.detectors))
	for _, d := range e.detectors {
		out = append(out, d.Detect(in)...)
	}
	rank(out)
	return out
}

// rank orders signals worst-first with a deterministic tiebreaker.
func rank(s []Signal) {
	sort.SliceStable(s, func(i, j int) bool {
		if a, b := s[i].Severity.Rank(), s[j].Severity.Rank(); a != b {
			return a > b
		}
		if s[i].Magnitude != s[j].Magnitude {
			return s[i].Magnitude > s[j].Magnitude
		}
		if s[i].FacilityID != s[j].FacilityID {
			return s[i].FacilityID < s[j].FacilityID
		}
		return s[i].Type < s[j].Type
	})
}

// ---- shared helpers ----

func groupByFacility(ms []metric.FacilityMetric) map[string][]metric.FacilityMetric {
	out := map[string][]metric.FacilityMetric{}
	for _, m := range ms {
		out[m.FacilityID] = append(out[m.FacilityID], m)
	}
	for id := range out {
		series := out[id]
		sort.Slice(series, func(i, j int) bool { return series[i].Date.Before(series[j].Date) })
	}
	return out
}

// splitWeeks returns the most recent window and the window before it.
func splitWeeks(ms []metric.FacilityMetric) (recent, prev []metric.FacilityMetric) {
	n := len(ms)
	w := min(7, n/2)
	if w == 0 {
		return nil, nil
	}
	return ms[n-w:], ms[n-2*w : n-w]
}

func sumRevenue(ms []metric.FacilityMetric) int64 {
	var t int64
	for _, m := range ms {
		t += m.Revenue.Pesewas()
	}
	return t
}

func sumUnbilled(ms []metric.FacilityMetric) int64 {
	var t int64
	for _, m := range ms {
		t += m.UnbilledAmount.Pesewas()
	}
	return t
}

func sumPatients(ms []metric.FacilityMetric) int {
	var t int
	for _, m := range ms {
		t += m.PatientsSeen
	}
	return t
}

func sumSubmitted(ms []metric.FacilityMetric) int {
	var t int
	for _, m := range ms {
		t += m.NHISClaimsSubmitted
	}
	return t
}

func sumDenied(ms []metric.FacilityMetric) int {
	var t int
	for _, m := range ms {
		t += m.NHISClaimsDenied
	}
	return t
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func clamp01(v float64) float64 {
	return math.Min(1, math.Max(0, v))
}
