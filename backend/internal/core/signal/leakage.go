package signal

import (
	"fmt"

	"github.com/xcreativs/gigmann/internal/core/money"
	"github.com/xcreativs/gigmann/internal/core/severity"
)

// LeakageDetector flags services delivered but unbilled (spec §6.3 revenue leakage).
type LeakageDetector struct{ th Thresholds }

// NewLeakageDetector builds a LeakageDetector.
func NewLeakageDetector(th Thresholds) LeakageDetector { return LeakageDetector{th: th} }

// Name identifies the detector.
func (LeakageDetector) Name() string { return "leakage" }

// Detect flags cumulative unbilled revenue above the configured thresholds.
func (d LeakageDetector) Detect(in Input) []Signal {
	var out []Signal
	for fid, series := range groupByFacility(in.Metrics) {
		unbilled := sumUnbilled(series)
		if unbilled < d.th.UnbilledWatch {
			continue
		}
		sev := severity.Watch
		if unbilled >= d.th.UnbilledCritical {
			sev = severity.Critical
		}
		// Normalised impact (≈1.0 at the critical threshold) so it ranks fairly
		// against ratio-based signals within a severity — not by raw pesewas. Guard
		// the threshold so a zero/misconfigured critical can't divide by zero.
		intensity := 1.0
		if d.th.UnbilledCritical > 0 {
			intensity = float64(unbilled) / float64(d.th.UnbilledCritical)
		}
		out = append(out, Signal{
			Type:              "revenue_leakage",
			FacilityID:        fid,
			Severity:          sev,
			Magnitude:         intensity,
			Headline:          fmt.Sprintf("%s delivered but unbilled", money.FromPesewas(unbilled)),
			SupportingFigures: map[string]any{"unbilled_pesewas": unbilled},
		})
	}
	return out
}
