// Package localnarrator is a deterministic, no-AI Narrator used when no
// ANTHROPIC_API_KEY is configured. It renders the computed context directly so
// the Daily Brief works in dev/demo without calling Claude. Pure (no I/O).
package localnarrator

import (
	"context"
	"strings"
	"time"

	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/intel"
	"github.com/xcreativs/gigmann/internal/ports"
)

// Narrator renders a brief straight from the context, without an LLM.
type Narrator struct{}

// Compile-time guarantee that Narrator satisfies the port.
var (
	_ ports.Narrator = (*Narrator)(nil)
	_ ports.Answerer = (*Narrator)(nil)
)

// New builds a local narrator.
func New() *Narrator { return &Narrator{} }

// NarrateBrief deterministically renders the context as a Daily Brief.
func (Narrator) NarrateBrief(_ context.Context, c intel.Context) (brief.Brief, error) {
	items := make([]brief.Item, 0, len(c.Items))
	for _, it := range c.Items {
		items = append(items, brief.Item{
			Severity:         it.Severity,
			FacilityID:       it.FacilityID,
			Headline:         it.Headline,
			Explanation:      it.FacilityName,
			SuggestedActions: []string{"Why?", "Open facility"},
		})
	}
	return brief.New(brief.Brief{
		ID:          "brief-" + c.Date.Format(time.DateOnly),
		Date:        c.Date,
		Prose:       "Good morning, Sammy. " + c.Pulse.Headline + ".",
		Items:       items,
		GeneratedAt: c.Date,
		Model:       "local-deterministic",
	})
}

// Answer deterministically summarises the context (no LLM) for the no-key path.
func (Narrator) Answer(_ context.Context, _ string, c intel.Context) (intel.Answer, error) {
	var b strings.Builder
	b.WriteString("From the current signals — " + c.Pulse.Headline + ".")
	citations := make([]string, 0, len(c.Items))
	for _, it := range c.Items {
		b.WriteString(" " + it.FacilityName + ": " + it.Headline + ".")
		citations = append(citations, it.FacilityID)
	}
	return intel.Answer{Text: b.String(), Citations: citations}, nil
}
