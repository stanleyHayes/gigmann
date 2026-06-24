// Package app holds application use cases. It orchestrates domain logic via
// ports and is where authorization decisions live. It imports no adapters.
package app

import (
	"context"
	"fmt"

	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/ports"
)

// FacilityService is the use case for working with facilities.
type FacilityService struct {
	repo ports.FacilityRepository
}

// NewFacilityService wires a FacilityService to its repository port.
func NewFacilityService(repo ports.FacilityRepository) *FacilityService {
	return &FacilityService{repo: repo}
}

// List returns all facilities in the network.
func (s *FacilityService) List(ctx context.Context) ([]facility.Facility, error) {
	facilities, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("app: list facilities: %w", err)
	}
	return facilities, nil
}
