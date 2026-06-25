package signal

import (
	"fmt"

	"github.com/xcreativs/gigmann/internal/core/metric"
	"github.com/xcreativs/gigmann/internal/core/severity"
)

// ClaimsDetector flags NHIS denial spikes and submission gaps (spec §6.3 claims health).
type ClaimsDetector struct{ th Thresholds }

// NewClaimsDetector builds a ClaimsDetector.
func NewClaimsDetector(th Thresholds) ClaimsDetector { return ClaimsDetector{th: th} }

// Name identifies the detector.
func (ClaimsDetector) Name() string { return "claims" }

// Detect flags denial-rate spikes and the diagnostic "recorded but not submitted" gap.
func (d ClaimsDetector) Detect(in Input) []Signal {
	var out []Signal
	for fid, series := range groupByFacility(in.Metrics) {
		recent, prev := splitWeeks(series)
		if len(recent) == 0 {
			continue
		}
		out = append(out, d.denialSpike(fid, recent)...)
		out = append(out, d.submissionGap(fid, recent, prev)...)
	}
	return out
}

func (d ClaimsDetector) denialSpike(fid string, recent []metric.FacilityMetric) []Signal {
	sub := sumSubmitted(recent)
	if sub == 0 {
		return nil
	}
	denied := sumDenied(recent)
	rate := float64(denied) / float64(sub)
	if rate < d.th.DenialRateWatch {
		return nil
	}
	sev := severity.Watch
	if rate >= d.th.DenialRateCritical {
		sev = severity.Critical
	}
	return []Signal{{
		Type: "denial_spike", FacilityID: fid, Severity: sev, Magnitude: rate,
		Headline:          fmt.Sprintf("NHIS denial rate at %.0f%%", rate*100),
		SupportingFigures: map[string]any{"denial_rate": round2(rate), "submitted": sub, "denied": denied},
	}}
}

func (d ClaimsDetector) submissionGap(fid string, recent, prev []metric.FacilityMetric) []Signal {
	if len(prev) == 0 {
		return nil
	}
	prevPatients, prevSub := sumPatients(prev), sumSubmitted(prev)
	if prevPatients == 0 || prevSub == 0 {
		return nil
	}
	patientChange := float64(sumPatients(recent)-prevPatients) / float64(prevPatients)
	subChange := float64(sumSubmitted(recent)-prevSub) / float64(prevSub)
	// Demand roughly flat but submissions collapsed → claims recorded, not submitted.
	if patientChange <= -0.10 || subChange > -d.th.SubmissionGapWatch {
		return nil
	}
	return []Signal{{
		Type: "submission_gap", FacilityID: fid, Severity: severity.Critical, Magnitude: -subChange,
		Headline: "Claims recorded but not submitted — demand is flat",
		SupportingFigures: map[string]any{
			"patient_change_pct":    round2(patientChange),
			"submission_change_pct": round2(subChange),
		},
	}}
}
