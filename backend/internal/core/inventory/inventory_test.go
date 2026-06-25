package inventory_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	require.NoError(t, err)
	assert.Equal(t, "Malaria RDT kit", it.Name)
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
			_, err := inventory.New(it)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestStockProjection(t *testing.T) {
	it := valid()
	days, ok := it.DaysOfStock()
	require.True(t, ok)
	assert.InDelta(t, 5, days, 0.0001)
	assert.True(t, it.StockOutImminent())

	it.DailyBurn = 0
	_, ok = it.DaysOfStock()
	assert.False(t, ok)
	assert.False(t, it.StockOutImminent())
}
