// Package severity models the cockpit's status triad (good / watch / critical),
// used for facility health, alerts, and the network pulse (spec §5).
package severity

// Severity is an ordered health/urgency signal.
type Severity string

const (
	Good     Severity = "good"
	Watch    Severity = "watch"
	Critical Severity = "critical"
)

// Valid reports whether s is a known severity.
func (s Severity) Valid() bool {
	switch s {
	case Good, Watch, Critical:
		return true
	default:
		return false
	}
}

// Rank orders severities by urgency (higher is worse): good=0, watch=1, critical=2.
// Used to sort "worst first" in the brief and attention feed.
func (s Severity) Rank() int {
	switch s {
	case Critical:
		return 2
	case Watch:
		return 1
	default:
		return 0
	}
}
