package app

import (
	"context"
	"errors"

	"github.com/xcreativs/gigmann/internal/core/alert"
	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/inventory"
	"github.com/xcreativs/gigmann/internal/core/staff"
)

// ErrFacilityNotFound is returned when no facility matches the requested id.
var ErrFacilityNotFound = errors.New("app: facility not found")

// FacilityDetail is the drill-down read model for one facility.
type FacilityDetail struct {
	Facility  facility.Facility
	Inventory []inventory.Item
	Staff     []staff.Member
	Alerts    []alert.Alert
}

// FacilityDetailService assembles a facility's drill-down from the network's
// read models (in-memory for the demo).
type FacilityDetailService struct {
	facilities map[string]facility.Facility
	inventory  []inventory.Item
	staff      []staff.Member
	alerts     []alert.Alert
}

// NewFacilityDetailService indexes the facilities and keeps the sub-resources.
func NewFacilityDetailService(
	facilities []facility.Facility, inv []inventory.Item, stf []staff.Member, alerts []alert.Alert,
) *FacilityDetailService {
	byID := make(map[string]facility.Facility, len(facilities))
	for _, f := range facilities {
		byID[f.ID] = f
	}
	return &FacilityDetailService{facilities: byID, inventory: inv, staff: stf, alerts: alerts}
}

// Detail returns the facility and its inventory/staff/alerts, or ErrFacilityNotFound.
// Facility managers may only drill into their own facility (ErrForbidden otherwise);
// executives see the whole network.
func (s *FacilityDetailService) Detail(_ context.Context, p auth.Principal, id string) (FacilityDetail, error) {
	f, ok := s.facilities[id]
	if !ok {
		return FacilityDetail{}, ErrFacilityNotFound
	}
	if !p.CanAccessFacility(id) {
		return FacilityDetail{}, ErrForbidden
	}
	return FacilityDetail{
		Facility:  f,
		Inventory: byFacility(s.inventory, id, func(i inventory.Item) string { return i.FacilityID }),
		Staff:     byFacility(s.staff, id, func(m staff.Member) string { return m.FacilityID }),
		Alerts:    byFacility(s.alerts, id, func(a alert.Alert) string { return a.FacilityID }),
	}, nil
}

func byFacility[T any](items []T, id string, facilityID func(T) string) []T {
	out := make([]T, 0)
	for _, it := range items {
		if facilityID(it) == id {
			out = append(out, it)
		}
	}
	return out
}
