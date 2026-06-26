//go:build integration

package anthropic_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/anthropic"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/intel"
)

// TestNarrateBriefLive exercises the real Anthropic API. It is build-tagged
// `integration` and skips unless ANTHROPIC_API_KEY is set, so normal `go test`
// (and CI without a key) never makes a billed call.
func TestNarrateBriefLive(t *testing.T) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		t.Skip("ANTHROPIC_API_KEY not set; skipping live Claude narration test")
	}
	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		model = "claude-sonnet-4-6"
	}

	n := anthropic.NewNarrator(key, model)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	c := intel.Context{
		Date: time.Date(2026, 6, 26, 0, 0, 0, 0, time.UTC),
		Pulse: intel.PulseSummary{
			Severity: severity.Critical, CriticalCount: 1, WatchCount: 1,
			Headline: "Network under strain — 1 critical, 1 to watch across 12 facilities",
		},
		Items: []intel.Item{
			{
				FacilityID: "tafo-maternity", FacilityName: "Tafo Maternity & Child Health",
				Type: "submission_gap", Severity: severity.Critical,
				Headline: "Claims recorded but not submitted — demand is flat",
				Figures:  map[string]any{"unbilled_pesewas": 4200000, "days_unsubmitted": 6},
			},
			{
				FacilityID: "kasoa", FacilityName: "Kasoa Polyclinic",
				Type: "denial_spike", Severity: severity.Watch,
				Headline: "NHIS denial rate at 19%",
				Figures:  map[string]any{"denial_rate": 0.19},
			},
		},
	}

	b, err := n.NarrateBrief(ctx, c)
	require.NoError(t, err)
	assert.NotEmpty(t, b.Prose, "Claude should produce narrated prose")
	assert.NotEmpty(t, b.Model, "brief should record the model")
	require.NotEmpty(t, b.Items, "brief should narrate at least one item")

	// Grounding guardrail: the narrator may only narrate the SUPPLIED facilities,
	// never invent one that was not in the deterministic context.
	supplied := map[string]bool{"tafo-maternity": true, "kasoa": true}
	for _, it := range b.Items {
		assert.Truef(t, supplied[it.FacilityID], "narrator invented a facility not in context: %q", it.FacilityID)
		assert.NotEmpty(t, it.Headline)
	}
	t.Logf("LIVE brief — model=%s items=%d\nprose: %s", b.Model, len(b.Items), b.Prose)
	for _, it := range b.Items {
		t.Logf("  [%s] %s — %s", it.Severity, it.FacilityID, it.Headline)
	}
}
