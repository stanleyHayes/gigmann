package signal

import (
	"fmt"

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
		out = append(out, Signal{
			Type: "revenue_leakage", FacilityID: fid, Severity: sev, Magnitude: float64(unbilled),
			Headline:          fmt.Sprintf("GH₵ %d delivered but unbilled", unbilled/100),
			SupportingFigures: map[string]any{"unbilled_pesewas": unbilled},
		})
	}
	return out
}
