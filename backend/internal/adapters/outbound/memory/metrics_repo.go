package memory

import (
	"context"

	"github.com/xcreativs/gigmann/internal/core/metric"
	"github.com/xcreativs/gigmann/internal/ports"
)

// MetricsRepo is an in-memory ports.MetricsRepository over a fixed metric series.
type MetricsRepo struct {
	series []metric.FacilityMetric
}

// NewMetricsRepo creates a repository over the given metric series.
func NewMetricsRepo(series ...metric.FacilityMetric) *MetricsRepo {
	return &MetricsRepo{series: append([]metric.FacilityMetric{}, series...)}
}

var _ ports.MetricsRepository = (*MetricsRepo)(nil)

// ListNetwork returns a copy of the full metric series.
func (r *MetricsRepo) ListNetwork(_ context.Context) ([]metric.FacilityMetric, error) {
	out := make([]metric.FacilityMetric, len(r.series))
	copy(out, r.series)
	return out, nil
}
