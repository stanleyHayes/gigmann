package intel_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/payer"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/core/signal"
	"github.com/xcreativs/gigmann/internal/intel"
)

func mustFacility(t *testing.T, id, name string) facility.Facility {
	t.Helper()
	mix, err := payer.New(70, 25, 5)
	require.NoError(t, err)
	f, err := facility.New(facility.Params{
		ID: id, Name: name, Region: "Central", Town: "Town", Type: "OPD", Beds: 10,
		Lifecycle: facility.LifecycleActive, Health: severity.Good, ManagerName: "M", PayerMix: mix,
	})
	require.NoError(t, err)
	return f
}

func TestBuildContext(t *testing.T) {
	asOf := time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC)
	facs := []facility.Facility{mustFacility(t, "f1", "Tafo Maternity"), mustFacility(t, "f2", "Kasoa")}
	sigs := []signal.Signal{
		{Type: "submission_gap", FacilityID: "f1", Severity: severity.Critical, Headline: "claims not submitted", SupportingFigures: map[string]any{"x": 1}},
		{Type: "denial_spike", FacilityID: "f2", Severity: severity.Watch, Headline: "denials up"},
	}
	pulse := signal.Pulse{Severity: severity.Critical, CriticalCount: 1, WatchCount: 1, Headline: "Network under strain"}

	c := intel.BuildContext(asOf, facs, sigs, pulse, 0)

	require.Len(t, c.Items, 2)
	assert.Equal(t, "Tafo Maternity", c.Items[0].FacilityName)
	assert.Equal(t, severity.Critical, c.Items[0].Severity)
	assert.Equal(t, severity.Critical, c.Pulse.Severity)
	assert.Equal(t, 1, c.Pulse.CriticalCount)
}

func TestBuildContextTopN(t *testing.T) {
	sigs := []signal.Signal{
		{Type: "a", FacilityID: "f1", Severity: severity.Critical},
		{Type: "b", FacilityID: "f2", Severity: severity.Watch},
		{Type: "c", FacilityID: "f3", Severity: severity.Watch},
	}
	c := intel.BuildContext(time.Now(), nil, sigs, signal.Pulse{}, 2)
	assert.Len(t, c.Items, 2)
	assert.Equal(t, "a", c.Items[0].Type)
	// FacilityName is never empty — it falls back to the id when unresolved, so the
	// narrator never emits broken text like ": Headline".
	for i, it := range c.Items {
		assert.NotEmpty(t, it.FacilityName, "item %d FacilityName", i)
	}
	assert.Equal(t, "f1", c.Items[0].FacilityName, "unresolved name falls back to the id")
}

func TestBuildContextResolvesNames(t *testing.T) {
	facilities := []facility.Facility{{ID: "kasoa", Name: "Kasoa Polyclinic"}}
	sigs := []signal.Signal{{Type: "a", FacilityID: "kasoa", Severity: severity.Critical}}
	c := intel.BuildContext(time.Now(), facilities, sigs, signal.Pulse{}, 0)
	require.Len(t, c.Items, 1)
	assert.Equal(t, "Kasoa Polyclinic", c.Items[0].FacilityName)
}
