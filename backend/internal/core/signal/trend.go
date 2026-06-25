package signal

import (
	"fmt"

	"github.com/xcreativs/gigmann/internal/core/severity"
)

// TrendDetector flags week-over-week revenue drops (spec §6.3 trend & delta).
type TrendDetector struct{ th Thresholds }

// NewTrendDetector builds a TrendDetector.
func NewTrendDetector(th Thresholds) TrendDetector { return TrendDetector{th: th} }

// Name identifies the detector.
func (TrendDetector) Name() string { return "trend" }

// Detect flags facilities whose revenue fell beyond the configured thresholds.
func (d TrendDetector) Detect(in Input) []Signal {
	var out []Signal
	for fid, series := range groupByFacility(in.Metrics) {
		recent, prev := splitWeeks(series)
		if len(recent) == 0 || len(prev) == 0 {
			continue
		}
		prevRev := sumRevenue(prev)
		if prevRev <= 0 {
			continue
		}
		recentRev := sumRevenue(recent)
		delta := float64(recentRev-prevRev) / float64(prevRev)
		if delta > -d.th.RevenueDropWatch {
			continue
		}
		sev := severity.Watch
		if delta <= -d.th.RevenueDropCritical {
			sev = severity.Critical
		}
		out = append(out, Signal{
			Type:       "revenue_drop",
			FacilityID: fid,
			Severity:   sev,
			Magnitude:  -delta,
			Headline:   fmt.Sprintf("Revenue down %.0f%% week-over-week", -delta*100),
			SupportingFigures: map[string]any{
				"delta_pct":        round2(delta),
				"recent_pesewas":   recentRev,
				"previous_pesewas": prevRev,
			},
		})
	}
	return out
}
