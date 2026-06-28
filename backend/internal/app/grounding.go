package app

import (
	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/core/facility"
)

// Grounding guardrails: the deterministic engine owns the facts, and the AI only
// narrates them. These helpers drop any AI-produced reference to a facility that
// is not in the supplied context, so an invented facility can never reach the
// user (CLAUDE.md §1: "the AI never invents a figure").

// knownFacilityIDs is the allow-list of facility ids the AI may reference.
func knownFacilityIDs(facilities []facility.Facility) map[string]bool {
	ids := make(map[string]bool, len(facilities))
	for _, f := range facilities {
		ids[f.ID] = true
	}
	return ids
}

// groundCitations keeps only citations that name a known facility.
func groundCitations(citations []string, known map[string]bool) []string {
	out := make([]string, 0, len(citations))
	for _, c := range citations {
		if known[c] {
			out = append(out, c)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// groundBriefItems keeps only items whose facility id is empty (network-level) or
// a known facility — dropping any item that references an invented facility.
func groundBriefItems(items []brief.Item, known map[string]bool) []brief.Item {
	out := make([]brief.Item, 0, len(items))
	for _, it := range items {
		if it.FacilityID == "" || known[it.FacilityID] {
			out = append(out, it)
		}
	}
	return out
}
