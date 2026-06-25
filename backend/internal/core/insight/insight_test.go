package insight_test

import (
	"errors"
	"testing"

	"github.com/xcreativs/gigmann/internal/core/insight"
)

func valid() insight.Insight {
	return insight.Insight{
		ID: "i1", Type: "claims_health", FacilityID: "kasoa", Content: "Denial rate spiking on a coding issue.",
	}
}

func TestNewValid(t *testing.T) {
	i, err := insight.New(valid())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if i.Type != "claims_health" {
		t.Errorf("type not set: %q", i.Type)
	}
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
			if _, err := insight.New(i); !errors.Is(err, tt.wantErr) {
				t.Fatalf("want %v, got %v", tt.wantErr, err)
			}
		})
	}
}
