// Package inventory holds the InventoryItem entity used by stock-out detection (spec §6.3).
package inventory

import (
	"errors"
	"strings"

	"github.com/xcreativs/gigmann/internal/core/money"
)

// Item is a stocked consumable at a facility (e.g. malaria RDT kits).
type Item struct {
	ID           string
	FacilityID   string
	Name         string
	Category     string
	StockLevel   int
	DailyBurn    float64
	ReorderPoint int
	LeadTimeDays int
	UnitCost     money.Cedis
}

// Validation errors.
var (
	ErrEmptyID         = errors.New("inventory: id is required")
	ErrEmptyFacilityID = errors.New("inventory: facility_id is required")
	ErrEmptyName       = errors.New("inventory: name is required")
	ErrNegative        = errors.New("inventory: stock/lead-time must be >= 0")
	ErrBadBurn         = errors.New("inventory: daily_burn must be >= 0")
)

// New validates and returns an Item.
func New(it Item) (Item, error) {
	it.ID = strings.TrimSpace(it.ID)
	it.FacilityID = strings.TrimSpace(it.FacilityID)
	it.Name = strings.TrimSpace(it.Name)
	switch {
	case it.ID == "":
		return Item{}, ErrEmptyID
	case it.FacilityID == "":
		return Item{}, ErrEmptyFacilityID
	case it.Name == "":
		return Item{}, ErrEmptyName
	case it.StockLevel < 0 || it.LeadTimeDays < 0:
		return Item{}, ErrNegative
	case it.DailyBurn < 0:
		return Item{}, ErrBadBurn
	}
	return it, nil
}

// DaysOfStock returns how many days of stock remain at the current burn rate.
// ok is false when burn is zero (run-out cannot be projected).
func (it Item) DaysOfStock() (days float64, ok bool) {
	if it.DailyBurn <= 0 {
		return 0, false
	}
	return float64(it.StockLevel) / it.DailyBurn, true
}

// StockOutImminent reports whether stock will run out within the supplier lead time.
func (it Item) StockOutImminent() bool {
	days, ok := it.DaysOfStock()
	return ok && days < float64(it.LeadTimeDays)
}
