package signal

import (
	"fmt"

	"github.com/xcreativs/gigmann/internal/core/severity"
)

// StockOutDetector flags inventory that will run out within the supplier lead time (spec §6.3).
type StockOutDetector struct{}

// NewStockOutDetector builds a StockOutDetector.
func NewStockOutDetector() StockOutDetector { return StockOutDetector{} }

// Name identifies the detector.
func (StockOutDetector) Name() string { return "stockout" }

// Detect flags items projected to run out before resupply arrives.
func (StockOutDetector) Detect(in Input) []Signal {
	var out []Signal
	for _, it := range in.Inventory {
		if !it.StockOutImminent() {
			continue
		}
		days, _ := it.DaysOfStock()
		sev := severity.Watch
		if days < float64(it.LeadTimeDays)/2 {
			sev = severity.Critical
		}
		// Normalised impact (≈1.0 when stock is fully out, 0 at the lead-time
		// threshold) so it ranks fairly against ratio-based signals within a severity.
		intensity := 1.0
		if it.LeadTimeDays > 0 {
			intensity = (float64(it.LeadTimeDays) - days) / float64(it.LeadTimeDays)
		}
		out = append(out, Signal{
			Type: "stock_out", FacilityID: it.FacilityID, Severity: sev, Magnitude: intensity,
			Headline: fmt.Sprintf("%s runs out in ~%.0f days (lead time %dd)", it.Name, days, it.LeadTimeDays),
			SupportingFigures: map[string]any{
				"item": it.Name, "days_left": round2(days), "lead_time_days": it.LeadTimeDays,
			},
		})
	}
	return out
}
