package signal

import (
	"fmt"
	"time"

	"github.com/xcreativs/gigmann/internal/core/severity"
)

// StaffDetector flags licence expiries and attrition risk (spec §6.3 staff signals).
type StaffDetector struct{ th Thresholds }

// NewStaffDetector builds a StaffDetector.
func NewStaffDetector(th Thresholds) StaffDetector { return StaffDetector{th: th} }

// Name identifies the detector.
func (StaffDetector) Name() string { return "staff" }

// Detect flags staff with imminent licence expiry or elevated attrition risk.
func (d StaffDetector) Detect(in Input) []Signal {
	var out []Signal
	for _, m := range in.Staff {
		if m.LicenceExpiringWithin(in.AsOf, d.th.LicenceWindowDays) {
			out = append(out, Signal{
				Type: "licence_expiry", FacilityID: m.FacilityID, Severity: severity.Watch,
				Magnitude:         licenceExpiryMagnitude(m.LicenceExpiry.Sub(in.AsOf), d.th.LicenceWindowDays),
				Headline:          fmt.Sprintf("%s licence expiring within %d days", m.Role, d.th.LicenceWindowDays),
				SupportingFigures: map[string]any{"staff_id": m.ID, "expiry": m.LicenceExpiry.Format("2006-01-02")},
			})
		}
		if m.AttritionRisk >= d.th.AttritionWatch {
			out = append(out, Signal{
				Type: "attrition_risk", FacilityID: m.FacilityID, Severity: severity.Watch, Magnitude: m.AttritionRisk,
				Headline:          m.Role + " at elevated attrition risk",
				SupportingFigures: map[string]any{"staff_id": m.ID, "attrition_risk": m.AttritionRisk},
			})
		}
	}
	return out
}

// licenceExpiryMagnitude normalises how imminent a licence expiry is, within the
// alert window, so a sooner (or already-passed) expiry ranks above a distant one
// — mirroring the threshold-normalised magnitudes the other detectors use rather
// than a flat constant. Just entering the window ≈ 0; expiring today ≈ 1; already
// expired > 1.
func licenceExpiryMagnitude(until time.Duration, windowDays int) float64 {
	if windowDays <= 0 {
		return 1
	}
	daysUntil := until.Hours() / 24
	mag := (float64(windowDays) - daysUntil) / float64(windowDays)
	if mag < 0 {
		return 0
	}
	return mag
}
