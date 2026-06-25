package inventory_test

import (
	"errors"
	"testing"

	"github.com/xcreativs/gigmann/internal/core/inventory"
	"github.com/xcreativs/gigmann/internal/core/money"
)

func valid() inventory.Item {
	return inventory.Item{
		ID: "rdt-asokwa", FacilityID: "asokwa", Name: "Malaria RDT kit", Category: "reagent",
		StockLevel: 50, DailyBurn: 10, ReorderPoint: 80, LeadTimeDays: 7, UnitCost: money.FromCedis(12, 0),
	}
}

func TestNewValid(t *testing.T) {
	it, err := inventory.New(valid())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if it.Name != "Malaria RDT kit" {
		t.Errorf("name not set: %q", it.Name)
	}
}

func TestNewInvariants(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(it *inventory.Item)
		wantErr error
	}{
		{"empty id", func(it *inventory.Item) { it.ID = "" }, inventory.ErrEmptyID},
		{"empty facility", func(it *inventory.Item) { it.FacilityID = "" }, inventory.ErrEmptyFacilityID},
		{"empty name", func(it *inventory.Item) { it.Name = "  " }, inventory.ErrEmptyName},
		{"negative stock", func(it *inventory.Item) { it.StockLevel = -1 }, inventory.ErrNegative},
		{"negative burn", func(it *inventory.Item) { it.DailyBurn = -2 }, inventory.ErrBadBurn},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := valid()
			tt.mutate(&it)
			if _, err := inventory.New(it); !errors.Is(err, tt.wantErr) {
				t.Fatalf("want %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestStockProjection(t *testing.T) {
	it := valid() // 50 / 10 = 5 days, lead time 7 → imminent (the Asokwa story)
	days, ok := it.DaysOfStock()
	if !ok || days != 5 {
		t.Fatalf("want 5 days ok, got %v ok=%v", days, ok)
	}
	if !it.StockOutImminent() {
		t.Error("expected stock-out imminent (5 days < 7 lead)")
	}

	it.DailyBurn = 0
	if _, ok := it.DaysOfStock(); ok {
		t.Error("expected ok=false when burn is zero")
	}
	if it.StockOutImminent() {
		t.Error("no burn should not be imminent")
	}
}
