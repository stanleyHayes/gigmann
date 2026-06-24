// Package memory provides in-memory outbound adapters for local dev and tests.
package memory

import (
	"context"
	"sync"

	"github.com/xcreativs/gigmann/internal/core/facility"
)

// FacilityRepo is an in-memory implementation of ports.FacilityRepository.
type FacilityRepo struct {
	mu         sync.RWMutex
	facilities []facility.Facility
}

// NewFacilityRepo creates a repository optionally seeded with facilities.
func NewFacilityRepo(seed ...facility.Facility) *FacilityRepo {
	return &FacilityRepo{facilities: append([]facility.Facility{}, seed...)}
}

// List returns a copy of the stored facilities.
func (r *FacilityRepo) List(_ context.Context) ([]facility.Facility, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]facility.Facility, len(r.facilities))
	copy(out, r.facilities)
	return out, nil
}
