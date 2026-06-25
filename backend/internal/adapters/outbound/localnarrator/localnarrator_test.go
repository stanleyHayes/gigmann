package localnarrator_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/localnarrator"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/intel"
)

func TestNarrateBrief(t *testing.T) {
	c := intel.Context{
		Date:  time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC),
		Pulse: intel.PulseSummary{Severity: severity.Critical, Headline: "Network under strain"},
		Items: []intel.Item{
			{Severity: severity.Critical, FacilityID: "tafo-maternity", FacilityName: "Tafo", Headline: "Tafo needs you first"},
		},
	}
	b, err := localnarrator.New().NarrateBrief(context.Background(), c)
	require.NoError(t, err)
	require.Len(t, b.Items, 1)
	assert.Contains(t, b.Prose, "Network under strain")
	assert.Equal(t, severity.Critical, b.Items[0].Severity)
	assert.Equal(t, "local-deterministic", b.Model)
}
