package facility_test

import (
	"errors"
	"testing"

	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/payer"
	"github.com/xcreativs/gigmann/internal/core/severity"
)

func validParams(t *testing.T) facility.Params {
	t.Helper()
	mix, err := payer.New(65, 25, 10)
	if err != nil {
		t.Fatalf("payer: %v", err)
	}
	return facility.Params{
		ID:          "f1",
		Name:        "Kasoa Polyclinic",
		Region:      "Central",
		Town:        "Kasoa",
		Type:        "OPD",
		Beds:        40,
		Lifecycle:   facility.LifecycleActive,
		Health:      severity.Good,
		ManagerName: "Ama Owusu",
		PayerMix:    mix,
	}
}

func TestNewValid(t *testing.T) {
	f, err := facility.New(validParams(t))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if f.ID != "f1" || f.Name != "Kasoa Polyclinic" || f.Health != severity.Good {
		t.Errorf("fields not set correctly: %+v", f)
	}
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
			if _, err := facility.New(p); !errors.Is(err, tt.wantErr) {
				t.Fatalf("want %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestLifecycleValid(t *testing.T) {
	for _, l := range []facility.Lifecycle{facility.LifecycleActive, facility.LifecycleRamping, facility.LifecycleFlagship} {
		if !l.Valid() {
			t.Errorf("%q should be valid", l)
		}
	}
	if facility.Lifecycle("x").Valid() {
		t.Error("unknown lifecycle reported valid")
	}
}
