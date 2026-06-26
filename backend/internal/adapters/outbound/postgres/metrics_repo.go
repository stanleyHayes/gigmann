package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/postgres/sqlcgen"
	"github.com/xcreativs/gigmann/internal/core/metric"
	"github.com/xcreativs/gigmann/internal/core/money"
	"github.com/xcreativs/gigmann/internal/ports"
)

// MetricsRepo is a PostgreSQL implementation of ports.MetricsRepository. The KPI
// numbers are always computed in Go (kpi.Compute) from the raw series this repo
// returns — Postgres is storage + efficient retrieval, never a source of figures.
type MetricsRepo struct {
	pool *pgxpool.Pool
	q    *sqlcgen.Queries
}

var _ ports.MetricsRepository = (*MetricsRepo)(nil)

// NewMetricsRepo builds a MetricsRepo over a pgx pool (the materialized-view
// helpers use the pool directly, so a bare DBTX is not enough).
func NewMetricsRepo(pool *pgxpool.Pool) *MetricsRepo {
	return &MetricsRepo{pool: pool, q: sqlcgen.New(pool)}
}

// ListNetwork returns the full metric series for the network (all facilities),
// ordered by (facility_id, metric_date) — backed by the primary key.
func (r *MetricsRepo) ListNetwork(ctx context.Context) ([]metric.FacilityMetric, error) {
	rows, err := r.q.ListNetworkMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres: list network metrics: %w", err)
	}
	return metricsFromModels(rows)
}

// ListFacilitySince returns one facility's metrics from `since` forward, ordered
// by date — the trailing-window / week-over-week access pattern served by the
// (facility_id, metric_date DESC) index.
func (r *MetricsRepo) ListFacilitySince(ctx context.Context, facilityID string, since time.Time) ([]metric.FacilityMetric, error) {
	rows, err := r.q.ListFacilityMetricsSince(ctx, sqlcgen.ListFacilityMetricsSinceParams{
		FacilityID: facilityID,
		MetricDate: dateToPg(since),
	})
	if err != nil {
		return nil, fmt.Errorf("postgres: list facility %q metrics: %w", facilityID, err)
	}
	return metricsFromModels(rows)
}

// Insert upserts a single facility-day metric (used by first-run seeding).
func (r *MetricsRepo) Insert(ctx context.Context, m metric.FacilityMetric) error {
	if err := r.q.InsertFacilityMetric(ctx, metricParams(m)); err != nil {
		return fmt.Errorf("postgres: insert metric %s/%s: %w", m.FacilityID, m.Date.Format("2006-01-02"), err)
	}
	return nil
}

// NetworkDailyRow is a row of the network_daily_metrics materialized view: a raw
// daily network rollup for charts (not a KPI figure).
type NetworkDailyRow struct {
	Date            time.Time
	Revenue         money.Cedis
	PatientsSeen    int
	Admissions      int
	NHISOutstanding money.Cedis
	UnbilledAmount  money.Cedis
}

// RefreshNetworkDaily recomputes the network_daily_metrics materialized view.
// (The cron worker, GEC-71, will call this on a schedule.)
func (r *MetricsRepo) RefreshNetworkDaily(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, `REFRESH MATERIALIZED VIEW network_daily_metrics`); err != nil {
		return fmt.Errorf("postgres: refresh network_daily_metrics: %w", err)
	}
	return nil
}

// NetworkDaily reads the materialized daily network rollup, oldest day first.
func (r *MetricsRepo) NetworkDaily(ctx context.Context) ([]NetworkDailyRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT metric_date, revenue, patients_seen, admissions, nhis_outstanding, unbilled_amount
		 FROM network_daily_metrics ORDER BY metric_date`)
	if err != nil {
		return nil, fmt.Errorf("postgres: read network_daily_metrics: %w", err)
	}
	defer rows.Close()
	var out []NetworkDailyRow
	for rows.Next() {
		var (
			d                                  time.Time
			revenue, nhisOutstanding, unbilled int64
			patients, admissions               int64
		)
		if err := rows.Scan(&d, &revenue, &patients, &admissions, &nhisOutstanding, &unbilled); err != nil {
			return nil, fmt.Errorf("postgres: scan network_daily_metrics: %w", err)
		}
		out = append(out, NetworkDailyRow{
			Date:            d.UTC(),
			Revenue:         money.FromPesewas(revenue),
			PatientsSeen:    int(patients),
			Admissions:      int(admissions),
			NHISOutstanding: money.FromPesewas(nhisOutstanding),
			UnbilledAmount:  money.FromPesewas(unbilled),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: iterate network_daily_metrics: %w", err)
	}
	return out, nil
}

func metricsFromModels(rows []sqlcgen.FacilityMetric) ([]metric.FacilityMetric, error) {
	out := make([]metric.FacilityMetric, 0, len(rows))
	for _, m := range rows {
		dm, err := metricFromModel(m)
		if err != nil {
			return nil, fmt.Errorf("postgres: map metric %s/%v: %w", m.FacilityID, m.MetricDate, err)
		}
		out = append(out, dm)
	}
	return out, nil
}

func metricFromModel(m sqlcgen.FacilityMetric) (metric.FacilityMetric, error) {
	return metric.New(metric.FacilityMetric{
		FacilityID:          m.FacilityID,
		Date:                dateFromPg(m.MetricDate),
		Revenue:             money.FromPesewas(m.Revenue),
		CashRevenue:         money.FromPesewas(m.CashRevenue),
		MoMoRevenue:         money.FromPesewas(m.MomoRevenue),
		PatientsSeen:        int(m.PatientsSeen),
		Admissions:          int(m.Admissions),
		OccupancyRate:       m.OccupancyRate,
		AvgWaitMinutes:      int(m.AvgWaitMinutes),
		NHISClaimsSubmitted: int(m.NhisClaimsSubmitted),
		NHISClaimsPaid:      int(m.NhisClaimsPaid),
		NHISClaimsDenied:    int(m.NhisClaimsDenied),
		NHISOutstanding:     money.FromPesewas(m.NhisOutstanding),
		UnbilledAmount:      money.FromPesewas(m.UnbilledAmount),
	})
}

func metricParams(m metric.FacilityMetric) sqlcgen.InsertFacilityMetricParams {
	return sqlcgen.InsertFacilityMetricParams{
		FacilityID:          m.FacilityID,
		MetricDate:          dateToPg(m.Date),
		Revenue:             m.Revenue.Pesewas(),
		CashRevenue:         m.CashRevenue.Pesewas(),
		MomoRevenue:         m.MoMoRevenue.Pesewas(),
		PatientsSeen:        i32(m.PatientsSeen),
		Admissions:          i32(m.Admissions),
		OccupancyRate:       m.OccupancyRate,
		AvgWaitMinutes:      i32(m.AvgWaitMinutes),
		NhisClaimsSubmitted: i32(m.NHISClaimsSubmitted),
		NhisClaimsPaid:      i32(m.NHISClaimsPaid),
		NhisClaimsDenied:    i32(m.NHISClaimsDenied),
		NhisOutstanding:     m.NHISOutstanding.Pesewas(),
		UnbilledAmount:      m.UnbilledAmount.Pesewas(),
	}
}
