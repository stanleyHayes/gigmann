package anthropic //nolint:testpackage // white-box: exercises the unexported parseBrief

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/core/severity"
)

func TestParseBriefValid(t *testing.T) {
	raw := []byte(`{
		"prose": "Good morning, Sammy.",
		"items": [
			{"severity":"critical","facility_id":"tafo-maternity","headline":"Tafo needs you first",
			 "explanation":"Claims recorded but not submitted","suggested_actions":["Why?","Message the manager"]}
		]
	}`)
	meta := briefMeta{
		id: "brief-2026-06-09", model: "claude-sonnet-4-6",
		date:        time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC),
		generatedAt: time.Date(2026, 6, 9, 5, 0, 0, 0, time.UTC),
	}

	b, err := parseBrief(meta, raw)
	require.NoError(t, err)
	require.Len(t, b.Items, 1)
	assert.Equal(t, "brief-2026-06-09", b.ID)
	assert.Equal(t, severity.Critical, b.Items[0].Severity)
	assert.Equal(t, "Tafo needs you first", b.Items[0].Headline)
	assert.Len(t, b.Items[0].SuggestedActions, 2)
}

func TestParseBriefInvalidSeverity(t *testing.T) {
	raw := []byte(`{"prose":"hi","items":[{"severity":"meh","facility_id":"f","headline":"h"}]}`)
	_, err := parseBrief(briefMeta{id: "b", date: time.Now()}, raw)
	require.Error(t, err)
}

func TestParseBriefBadJSON(t *testing.T) {
	_, err := parseBrief(briefMeta{id: "b", date: time.Now()}, []byte("{not json"))
	require.Error(t, err)
}
