package postgres

import (
	"context"
	"fmt"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/postgres/sqlcgen"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/payer"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/ports"
)

// FacilityRepo is a PostgreSQL implementation of ports.FacilityRepository.
type FacilityRepo struct {
	q *sqlcgen.Queries
}

// Compile-time guarantee that FacilityRepo satisfies the port.
var _ ports.FacilityRepository = (*FacilityRepo)(nil)

// NewFacilityRepo builds a FacilityRepo over a pgx pool (or any sqlcgen.DBTX).
func NewFacilityRepo(db sqlcgen.DBTX) *FacilityRepo {
	return &FacilityRepo{q: sqlcgen.New(db)}
}

// List returns all facilities, mapped to the domain model.
func (r *FacilityRepo) List(ctx context.Context) ([]facility.Facility, error) {
	rows, err := r.q.ListFacilities(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres: list facilities: %w", err)
	}
	out := make([]facility.Facility, 0, len(rows))
	for _, row := range rows {
		f, ferr := facilityFromRow(row)
		if ferr != nil {
			return nil, fmt.Errorf("postgres: map facility %q: %w", row.ID, ferr)
		}
		out = append(out, f)
	}
	return out, nil
}

func facilityFromRow(row sqlcgen.ListFacilitiesRow) (facility.Facility, error) {
	mix, err := payer.New(int(row.PayerNhis), int(row.PayerCashMomo), int(row.PayerPrivate))
	if err != nil {
		return facility.Facility{}, err
	}
	return facility.New(facility.Params{
		ID:          row.ID,
		Name:        row.Name,
		Region:      facility.Region(row.Region),
		Town:        row.Town,
		Type:        row.Type,
		Beds:        int(row.Beds),
		Lifecycle:   facility.Lifecycle(row.Lifecycle),
		Health:      severity.Severity(row.Health),
		ManagerName: row.ManagerName,
		PayerMix:    mix,
		Latitude:    row.Latitude,
		Longitude:   row.Longitude,
	})
}
