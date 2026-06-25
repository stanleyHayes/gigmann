package insight_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/core/insight"
)

func valid() insight.Insight {
	return insight.Insight{
		ID: "i1", Type: "claims_health", FacilityID: "kasoa", Content: "Denial rate spiking on a coding issue.",
	}
}

func TestNewValid(t *testing.T) {
	i, err := insight.New(valid())
	require.NoError(t, err)
	assert.Equal(t, "claims_health", i.Type)
}

func TestNewInvariants(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(i *insight.Insight)
		wantErr error
	}{
		{"empty id", func(i *insight.Insight) { i.ID = "" }, insight.ErrEmptyID},
		{"empty type", func(i *insight.Insight) { i.Type = " " }, insight.ErrEmptyType},
		{"empty content", func(i *insight.Insight) { i.Content = "" }, insight.ErrEmptyContent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := valid()
			tt.mutate(&i)
			_, err := insight.New(i)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
