package signal_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/core/signal"
	"github.com/xcreativs/gigmann/internal/seed"
)

// TestEngineOverSyntheticNetwork proves the engine surfaces the planted stories
// from the deterministic generator, ranked worst-first and reproducibly.
func TestEngineOverSyntheticNetwork(t *testing.T) {
	asOf := time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC)
	net := seed.Generate(7, asOf, 14)
	in := signal.Input{
		AsOf: asOf, Facilities: net.Facilities, Metrics: net.Metrics,
		Inventory: net.Inventory, Staff: net.Staff,
	}
	eng := signal.Default(signal.DefaultThresholds())

	sigs := eng.Run(in)
	require.NotEmpty(t, sigs)

	// worst-first ordering
	assert.GreaterOrEqual(t, sigs[0].Severity.Rank(), sigs[len(sigs)-1].Severity.Rank())
	// deterministic
	assert.Equal(t, sigs, eng.Run(in))

	// planted stories surface
	assert.True(t, hasFacilityType(sigs, "tafo-maternity", "submission_gap"), "Tafo: claims recorded but not submitted")
	assert.True(t, hasFacilityType(sigs, "asokwa", "stock_out"), "Asokwa: RDT stock-out")
	assert.True(t, hasFacilityType(sigs, "kasoa", "denial_spike"), "Kasoa: NHIS denial spike")
	assert.True(t,
		hasFacilityType(sigs, "tamale-north", "licence_expiry") || hasFacilityType(sigs, "tamale-north", "attrition_risk"),
		"Tamale: staff signal")

	pulse := signal.NetworkPulse(net.Facilities, sigs)
	assert.Equal(t, severity.Critical, pulse.Severity)
	assert.Positive(t, pulse.CriticalCount)
}

func hasFacilityType(sigs []signal.Signal, fid, typ string) bool {
	for _, s := range sigs {
		if s.FacilityID == fid && s.Type == typ {
			return true
		}
	}
	return false
}
