// Package app holds application use cases. It orchestrates domain logic via
// ports and is where authorization decisions live. It imports no adapters.
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/metric"
	"github.com/xcreativs/gigmann/internal/core/money"
	"github.com/xcreativs/gigmann/internal/ports"
)

// FacilitySummary is the facility roster item plus latest deterministic
// operating figures for the Network tiles.
type FacilitySummary struct {
	Facility      facility.Facility
	LatestRevenue money.Cedis
	OccupancyRate float64
	PatientsSeen  int
	HasLatest     bool
}

// FacilityService is the use case for working with facilities.
type FacilityService struct {
	repo    ports.FacilityRepository
	metrics ports.MetricsRepository
}

// NewFacilityService wires a FacilityService to its repository port.
func NewFacilityService(repo ports.FacilityRepository, metrics ...ports.MetricsRepository) *FacilityService {
	var m ports.MetricsRepository
	if len(metrics) > 0 {
		m = metrics[0]
	}
	return &FacilityService{repo: repo, metrics: m}
}

// List returns all facilities in the network.
func (s *FacilityService) List(ctx context.Context) ([]facility.Facility, error) {
	facilities, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("app: list facilities: %w", err)
	}
	return facilities, nil
}

// ListSummaries returns all facilities with the latest raw metric row attached
// when the metrics repository is available. Figures are computed/read from raw
// data here, never guessed by the UI.
func (s *FacilityService) ListSummaries(ctx context.Context) ([]FacilitySummary, error) {
	facilities, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]FacilitySummary, 0, len(facilities))
	latest := map[string]metric.FacilityMetric{}
	if s.metrics != nil {
		series, err := s.metrics.ListNetwork(ctx)
		if err != nil {
			return nil, fmt.Errorf("app: list facility metrics: %w", err)
		}
		latest = latestMetricByFacility(series)
	}
	for _, f := range facilities {
		summary := FacilitySummary{Facility: f}
		if m, ok := latest[f.ID]; ok {
			summary.LatestRevenue = m.Revenue
			summary.OccupancyRate = m.OccupancyRate
			summary.PatientsSeen = m.PatientsSeen
			summary.HasLatest = true
		}
		out = append(out, summary)
	}
	return out, nil
}

func latestMetricByFacility(series []metric.FacilityMetric) map[string]metric.FacilityMetric {
	latest := map[string]metric.FacilityMetric{}
	dates := map[string]time.Time{}
	for _, m := range series {
		if seen, ok := dates[m.FacilityID]; !ok || m.Date.After(seen) {
			latest[m.FacilityID] = m
			dates[m.FacilityID] = m.Date
		}
	}
	return latest
}
