package app

import (
	"context"
	"fmt"

	"github.com/xcreativs/gigmann/internal/core/kpi"
	"github.com/xcreativs/gigmann/internal/ports"
)

// MetricsService is the use case for the deterministic network KPI view. It loads
// the raw metric series from the repository and computes KPIs in Go (kpi.Compute);
// no figure ever originates in SQL.
type MetricsService struct {
	metrics ports.MetricsRepository
}

// NewMetricsService wires a MetricsService over a metrics repository.
func NewMetricsService(metrics ports.MetricsRepository) *MetricsService {
	return &MetricsService{metrics: metrics}
}

// Network returns the computed network KPIs.
func (s *MetricsService) Network(ctx context.Context) (kpi.Network, error) {
	series, err := s.metrics.ListNetwork(ctx)
	if err != nil {
		return kpi.Network{}, fmt.Errorf("metrics: load network series: %w", err)
	}
	return kpi.Compute(series), nil
}
