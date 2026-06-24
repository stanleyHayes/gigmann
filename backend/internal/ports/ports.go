// Package ports declares the interfaces the application layer depends on.
// Outbound adapters (Postgres, Redis, in-memory, Anthropic) implement these.
package ports

import (
	"context"

	"github.com/xcreativs/gigmann/internal/core/facility"
)

// FacilityRepository is a driven port for reading/writing facilities.
type FacilityRepository interface {
	List(ctx context.Context) ([]facility.Facility, error)
}
