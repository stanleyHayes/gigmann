package signal

import (
	"fmt"

	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/severity"
)

// Pulse is the composite, AI-assessed health of the whole network (spec §5.2, §6.3).
type Pulse struct {
	Score         float64 // 0..1, higher is healthier
	Severity      severity.Severity
	CriticalCount int
	WatchCount    int
	Headline      string
}

const (
	criticalWeight = 0.15
	watchWeight    = 0.05
)

// NetworkPulse derives the composite from the active signals and facility count.
func NetworkPulse(facilities []facility.Facility, signals []Signal) Pulse {
	var crit, watch int
	for _, s := range signals {
		switch s.Severity {
		case severity.Critical:
			crit++
		case severity.Watch:
			watch++
		case severity.Good:
			// healthy signal, not counted against the pulse
		}
	}
	score := clamp01(1.0 - (float64(crit)*criticalWeight + float64(watch)*watchWeight))
	sev := severity.Good
	switch {
	case crit > 0:
		sev = severity.Critical
	case watch > 0:
		sev = severity.Watch
	}
	return Pulse{
		Score: round2(score), Severity: sev, CriticalCount: crit, WatchCount: watch,
		Headline: fmt.Sprintf("Network %s — %d critical, %d to watch across %d facilities",
			pulseWord(sev), crit, watch, len(facilities)),
	}
}

func pulseWord(s severity.Severity) string {
	if s == severity.Critical {
		return "under strain"
	}
	if s == severity.Watch {
		return "needs attention"
	}
	return "steady"
}
