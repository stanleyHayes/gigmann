package app //nolint:testpackage // white-box: exercises the unexported grounding guards

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/core/facility"
)

func TestGroundCitationsDropsInvented(t *testing.T) {
	known := knownFacilityIDs([]facility.Facility{{ID: "kasoa"}, {ID: "nima"}})

	assert.Equal(t, []string{"kasoa", "nima"}, groundCitations([]string{"kasoa", "invented", "nima"}, known))
	assert.Nil(t, groundCitations([]string{"made-up"}, known), "all-invented → nil, not a misleading citation")
	assert.Nil(t, groundCitations(nil, known))
}

func TestGroundBriefItemsDropsInventedFacility(t *testing.T) {
	known := knownFacilityIDs([]facility.Facility{{ID: "kasoa"}, {ID: "nima"}})
	items := []brief.Item{
		{FacilityID: "kasoa"},
		{FacilityID: ""}, // network-level item is kept
		{FacilityID: "ghost-facility"},
		{FacilityID: "nima"},
	}
	got := groundBriefItems(items, known)
	assert.Len(t, got, 3)
	for _, it := range got {
		assert.NotEqual(t, "ghost-facility", it.FacilityID, "invented facility dropped")
	}
}
