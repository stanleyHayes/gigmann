// Package intel assembles the deterministic context handed to the AI narrator
// (spec §6.2): computed signals + facility facts, ranked and trimmed. No I/O.
package intel

import (
	"time"

	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/core/signal"
)

// PulseSummary is the network-level health summary for the brief header.
type PulseSummary struct {
	Severity      severity.Severity `json:"severity"`
	CriticalCount int               `json:"critical_count"`
	WatchCount    int               `json:"watch_count"`
	Headline      string            `json:"headline"`
}

// Item is one prioritised signal with the facts the narrator may cite.
type Item struct {
	FacilityID   string            `json:"facility_id"`
	FacilityName string            `json:"facility_name"`
	Type         string            `json:"type"`
	Severity     severity.Severity `json:"severity"`
	Headline     string            `json:"headline"`
	Figures      map[string]any    `json:"figures,omitempty"`
}

// Context is the structured input to a Narrator. It is fully computed: the AI
// narrates these facts and must not invent figures beyond them.
type Context struct {
	Date  time.Time    `json:"date"`
	Pulse PulseSummary `json:"pulse"`
	Items []Item       `json:"items"`
}

// Answer is a grounded response to a natural-language question about the network.
type Answer struct {
	Text      string   `json:"text"`
	Citations []string `json:"citations,omitempty"`
}

// BuildContext maps ranked signals to context items (resolving facility names)
// and trims to the top N (0 = keep all). Order is preserved from the engine.
func BuildContext(asOf time.Time, facilities []facility.Facility, signals []signal.Signal, pulse signal.Pulse, topN int) Context {
	names := make(map[string]string, len(facilities))
	for _, f := range facilities {
		names[f.ID] = f.Name
	}

	items := make([]Item, 0, len(signals))
	for _, s := range signals {
		// Fall back to the id if the name can't be resolved, so FacilityName is
		// never empty (which would produce broken narration like ": Headline").
		name := names[s.FacilityID]
		if name == "" {
			name = s.FacilityID
		}
		items = append(items, Item{
			FacilityID:   s.FacilityID,
			FacilityName: name,
			Type:         s.Type,
			Severity:     s.Severity,
			Headline:     s.Headline,
			Figures:      s.SupportingFigures,
		})
	}
	if topN > 0 && len(items) > topN {
		items = items[:topN]
	}

	return Context{
		Date: asOf,
		Pulse: PulseSummary{
			Severity:      pulse.Severity,
			CriticalCount: pulse.CriticalCount,
			WatchCount:    pulse.WatchCount,
			Headline:      pulse.Headline,
		},
		Items: items,
	}
}
