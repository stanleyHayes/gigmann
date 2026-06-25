// Package kpi computes deterministic network KPIs and week-over-week deltas from
// the facility metric series (spec §6.3). Pure domain code: every number here is
// computed in code — the AI never produces a KPI figure. Mirrors the signal
// engine's week split so the brief and the KPI screen agree.
package kpi

import (
	"math"
	"sort"
	"time"

	"github.com/xcreativs/gigmann/internal/core/metric"
)

// Direction is the sign of a week-over-week change.
type Direction string

// Direction values.
const (
	Up   Direction = "up"
	Down Direction = "down"
	Flat Direction = "flat"
)

const eps = 1e-9

// Point is one day's network-aggregate value for a KPI.
type Point struct {
	Date  time.Time
	Value float64
}

// KPI is a single network metric with its daily series and week-over-week delta.
// Unit tells the client how to interpret the values: "pesewas" (minor money
// units), "ratio" (0..1), or "count". HigherIsBetter lets the client colour the
// delta by business meaning (e.g. a rising denial rate is bad).
type KPI struct {
	Key            string
	Label          string
	Unit           string
	HigherIsBetter bool
	Current        float64
	Previous       float64
	DeltaPct       float64
	Direction      Direction
	Series         []Point
}

// Network is the full set of computed KPIs as of the latest metric date.
type Network struct {
	AsOf time.Time
	KPIs []KPI
}

// dayAgg is the network aggregate for a single day across all facilities.
type dayAgg struct {
	date      time.Time
	revenue   int64 // pesewas
	patients  int
	submitted int
	denied    int
	occSum    float64
	occN      int
}

// Compute aggregates the metric series by day and derives the network KPIs.
func Compute(metrics []metric.FacilityMetric) Network {
	days := aggregate(metrics)
	var asOf time.Time
	if len(days) > 0 {
		asOf = days[len(days)-1].date
	}
	recent, prev := window(days)

	return Network{
		AsOf: asOf,
		KPIs: []KPI{
			buildSum("revenue", "Network revenue", "pesewas", true, days, recent, prev,
				func(a dayAgg) float64 { return float64(a.revenue) }),
			buildSum("patients", "Patients seen", "count", true, days, recent, prev,
				func(a dayAgg) float64 { return float64(a.patients) }),
			buildDenial(days, recent, prev),
			buildOccupancy(days, recent, prev),
		},
	}
}

func aggregate(metrics []metric.FacilityMetric) []dayAgg {
	byDay := map[int64]*dayAgg{}
	for _, m := range metrics {
		d := m.Date.UTC().Truncate(24 * time.Hour)
		a := byDay[d.Unix()]
		if a == nil {
			a = &dayAgg{date: d}
			byDay[d.Unix()] = a
		}
		a.revenue += m.Revenue.Pesewas()
		a.patients += m.PatientsSeen
		a.submitted += m.NHISClaimsSubmitted
		a.denied += m.NHISClaimsDenied
		a.occSum += m.OccupancyRate
		a.occN++
	}
	out := make([]dayAgg, 0, len(byDay))
	for _, a := range byDay {
		out = append(out, *a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].date.Before(out[j].date) })
	return out
}

// window returns the most recent block of days and the block before it.
func window(days []dayAgg) (recent, prev []dayAgg) {
	n := len(days)
	w := min(7, n/2)
	if w == 0 {
		return nil, nil
	}
	return days[n-w:], days[n-2*w : n-w]
}

func buildSum(key, label, unit string, higher bool, days, recent, prev []dayAgg, val func(dayAgg) float64) KPI {
	cur := sumOver(recent, val)
	pre := sumOver(prev, val)
	return KPI{
		Key: key, Label: label, Unit: unit, HigherIsBetter: higher,
		Current: round2(cur), Previous: round2(pre), DeltaPct: deltaPct(cur, pre),
		Direction: direction(cur, pre), Series: series(days, val),
	}
}

func buildDenial(days, recent, prev []dayAgg) KPI {
	rate := func(den, sub int) float64 {
		if sub == 0 {
			return 0
		}
		return float64(den) / float64(sub)
	}
	dayVal := func(a dayAgg) float64 { return rate(a.denied, a.submitted) }
	winRate := func(w []dayAgg) float64 {
		var den, sub int
		for _, a := range w {
			den += a.denied
			sub += a.submitted
		}
		return rate(den, sub)
	}
	cur, pre := winRate(recent), winRate(prev)
	return KPI{
		Key: "denial_rate", Label: "NHIS denial rate", Unit: "ratio", HigherIsBetter: false,
		Current: round2(cur), Previous: round2(pre), DeltaPct: deltaPct(cur, pre),
		Direction: direction(cur, pre), Series: series(days, dayVal),
	}
}

func buildOccupancy(days, recent, prev []dayAgg) KPI {
	dayVal := func(a dayAgg) float64 {
		if a.occN == 0 {
			return 0
		}
		return a.occSum / float64(a.occN)
	}
	winMean := func(w []dayAgg) float64 {
		var s float64
		var n int
		for _, a := range w {
			s += a.occSum
			n += a.occN
		}
		if n == 0 {
			return 0
		}
		return s / float64(n)
	}
	cur, pre := winMean(recent), winMean(prev)
	return KPI{
		Key: "occupancy", Label: "Bed occupancy", Unit: "ratio", HigherIsBetter: true,
		Current: round2(cur), Previous: round2(pre), DeltaPct: deltaPct(cur, pre),
		Direction: direction(cur, pre), Series: series(days, dayVal),
	}
}

func series(days []dayAgg, val func(dayAgg) float64) []Point {
	pts := make([]Point, 0, len(days))
	for _, a := range days {
		pts = append(pts, Point{Date: a.date, Value: round2(val(a))})
	}
	return pts
}

func sumOver(w []dayAgg, val func(dayAgg) float64) float64 {
	var t float64
	for _, a := range w {
		t += val(a)
	}
	return t
}

func deltaPct(cur, prev float64) float64 {
	if prev == 0 {
		return 0
	}
	return round2((cur - prev) / prev)
}

func direction(cur, prev float64) Direction {
	switch {
	case cur > prev+eps:
		return Up
	case cur < prev-eps:
		return Down
	default:
		return Flat
	}
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }
