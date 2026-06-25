package app

import (
	"context"

	"github.com/xcreativs/gigmann/internal/core/kpi"
	"github.com/xcreativs/gigmann/internal/core/metric"
)

// MetricsService is the use case for the deterministic network KPI view.
type MetricsService struct {
	metrics []metric.FacilityMetric
}

// NewMetricsService wires a MetricsService over the network's metric series.
func NewMetricsService(metrics []metric.FacilityMetric) *MetricsService {
	return &MetricsService{metrics: metrics}
}

// Network returns the computed network KPIs.
func (s *MetricsService) Network(_ context.Context) (kpi.Network, error) {
	return kpi.Compute(s.metrics), nil
}
