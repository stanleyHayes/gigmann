package brief_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/core/severity"
)

func valid() brief.Brief {
	return brief.Brief{
		ID: "b-2026-06-09", Date: time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC),
		Prose: "Good morning, Sammy.",
		Items: []brief.Item{
			{Severity: severity.Critical, FacilityID: "tafo-maternity", Headline: "Tafo Maternity needs you first",
				Explanation: "Revenue down 22% — claims not submitted", SuggestedActions: []string{"Why?", "Message the manager"}},
		},
		Model: "claude-sonnet-4-6",
	}
}

func TestNewValid(t *testing.T) {
	b, err := brief.New(valid())
	require.NoError(t, err)
	require.Len(t, b.Items, 1)
	assert.Equal(t, severity.Critical, b.Items[0].Severity)
}

func TestNewInvariants(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(b *brief.Brief)
		wantErr error
	}{
		{"empty id", func(b *brief.Brief) { b.ID = "" }, brief.ErrEmptyID},
		{"zero date", func(b *brief.Brief) { b.Date = time.Time{} }, brief.ErrZeroDate},
		{"bad item severity", func(b *brief.Brief) { b.Items[0].Severity = "meh" }, brief.ErrInvalidItemSeverity},
		{"empty headline", func(b *brief.Brief) { b.Items[0].Headline = "  " }, brief.ErrEmptyHeadline},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := valid()
			tt.mutate(&b)
			_, err := brief.New(b)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
