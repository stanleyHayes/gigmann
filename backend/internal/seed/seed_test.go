package seed_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/seed"
)

var asOf = time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC)

func TestGenerateDeterministic(t *testing.T) {
	a := seed.Generate(7, asOf, 14)
	b := seed.Generate(7, asOf, 14)
	require.Equal(t, a, b, "same seed + as-of must reproduce an identical network")
}

func TestNetworkShape(t *testing.T) {
	net := seed.Generate(7, asOf, 14)
	require.Len(t, net.Facilities, 12)
	assert.Len(t, net.Metrics, 12*14)
	assert.Len(t, net.Approvals, 3)
	assert.NotEmpty(t, net.Alerts)
	assert.NotEmpty(t, net.Inventory)
	for _, f := range net.Facilities {
		assert.True(t, f.PayerMix.Valid(), f.ID)
	}
}

func TestGhanaGrounding(t *testing.T) {
	net := seed.Generate(7, asOf, 14)
	ids := map[string]bool{}
	for _, f := range net.Facilities {
		ids[f.ID] = true
	}
	for _, want := range []string{"assin-fosu", "tafo-maternity", "kasoa", "adansi", "asokwa", "tamale-north"} {
		assert.Contains(t, ids, want)
	}
}

func TestPlantedStories(t *testing.T) {
	net := seed.Generate(7, asOf, 14)

	assert.Equal(t, severity.Critical, findFacility(t, net, "tafo-maternity").Health)
	assert.True(t, hasAlert(net, "tafo-maternity", "revenue_drop"))
	assert.Equal(t, severity.Good, findFacility(t, net, "adansi").Health)

	var asokwaImminent bool
	for _, it := range net.Inventory {
		if it.FacilityID == "asokwa" && it.StockOutImminent() {
			asokwaImminent = true
		}
	}
	assert.True(t, asokwaImminent, "Asokwa RDT should be stock-out imminent")

	require.Len(t, net.Approvals, 3)
	var ultrasound bool
	for _, a := range net.Approvals {
		if a.ID == "ap-ultrasound" {
			ultrasound = true
			assert.Equal(t, int64(8500000), a.Amount.Pesewas())
		}
	}
	assert.True(t, ultrasound, "the GH₵ 85,000 ultrasound approval should be present")
}

func TestTafoUnbilledGrows(t *testing.T) {
	net := seed.Generate(7, asOf, 14)
	var earliest, latest int64
	var seen bool
	for _, m := range net.Metrics {
		if m.FacilityID != "tafo-maternity" {
			continue
		}
		if !seen {
			earliest = m.UnbilledAmount.Pesewas()
			seen = true
		}
		latest = m.UnbilledAmount.Pesewas()
	}
	require.True(t, seen)
	assert.Greater(t, latest, earliest, "Tafo unbilled amount should grow across the window")
}

func findFacility(t *testing.T, net seed.Network, id string) facility.Facility {
	t.Helper()
	for _, f := range net.Facilities {
		if f.ID == id {
			return f
		}
	}
	t.Fatalf("facility %q not found", id)
	return facility.Facility{}
}

func hasAlert(net seed.Network, facilityID, alertType string) bool {
	for _, a := range net.Alerts {
		if a.FacilityID == facilityID && a.Type == alertType {
			return true
		}
	}
	return false
}
