package facility_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/payer"
	"github.com/xcreativs/gigmann/internal/core/severity"
)

func validParams(t *testing.T) facility.Params {
	t.Helper()
	mix, err := payer.New(65, 25, 10)
	require.NoError(t, err)
	return facility.Params{
		ID: "f1", Name: "Kasoa Polyclinic", Region: "Central", Town: "Kasoa",
		Type: "OPD", Beds: 40, Lifecycle: facility.LifecycleActive, Health: severity.Good,
		ManagerName: "Ama Owusu", PayerMix: mix,
	}
}

func TestNewValid(t *testing.T) {
	f, err := facility.New(validParams(t))
	require.NoError(t, err)
	assert.Equal(t, "f1", f.ID)
	assert.Equal(t, severity.Good, f.Health)
}

func TestNewInvariants(t *testing.T) {
	badMix := payer.Mix{NHIS: 10, CashMoMo: 10, Private: 10}
	tests := []struct {
		name    string
		mutate  func(p *facility.Params)
		wantErr error
	}{
		{"empty id", func(p *facility.Params) { p.ID = "  " }, facility.ErrEmptyID},
		{"empty name", func(p *facility.Params) { p.Name = "" }, facility.ErrEmptyName},
		{"empty region", func(p *facility.Params) { p.Region = " " }, facility.ErrEmptyRegion},
		{"negative beds", func(p *facility.Params) { p.Beds = -1 }, facility.ErrNegativeBeds},
		{"bad lifecycle", func(p *facility.Params) { p.Lifecycle = "zombie" }, facility.ErrInvalidLifecycle},
		{"bad health", func(p *facility.Params) { p.Health = "meh" }, facility.ErrInvalidHealth},
		{"bad payer mix", func(p *facility.Params) { p.PayerMix = badMix }, facility.ErrInvalidPayerMix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validParams(t)
			tt.mutate(&p)
			_, err := facility.New(p)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestLifecycleValid(t *testing.T) {
	for _, l := range []facility.Lifecycle{facility.LifecycleActive, facility.LifecycleRamping, facility.LifecycleFlagship} {
		assert.Truef(t, l.Valid(), "%q should be valid", l)
	}
	assert.False(t, facility.Lifecycle("x").Valid())
}
