package signal

import (
	"fmt"

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
				Type: "licence_expiry", FacilityID: m.FacilityID, Severity: severity.Watch, Magnitude: 1,
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
