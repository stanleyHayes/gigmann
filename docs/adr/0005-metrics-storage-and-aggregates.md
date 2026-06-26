# ADR-0005: Metrics storage & aggregates on native Postgres

- **Status:** accepted
- **Date:** 2026-06-26
- **Deciders:** Owner (Stanley) + engineering

## Context
GEC-12 needs fast week-over-week / trailing-window metric queries and common
aggregates **on Render's managed Postgres**, which has no TimescaleDB (ADR-0001,
OQ-4). The product's defining rule (ADR-0004, CLAUDE.md §1) is that **every KPI is
computed by the Go signal engine (`kpi.Compute`)** — the database must never be a
second source of figures.

## Decision
- **Raw storage, Go computation.** `facility_metrics` keeps the per-facility-day
  series (PK `(facility_id, metric_date)`, plus index
  `idx_facility_metrics_facility_date (facility_id, metric_date DESC)`).
  `ports.MetricsRepository.ListNetwork` returns the raw series and
  `MetricsService` feeds it to `kpi.Compute` — **no KPI originates in SQL**. A
  `ListFacilitySince` adapter method serves the trailing-window / WoW pattern off
  the same index.
- **Materialized view for charts/scale, not for truth.** `network_daily_metrics`
  is a raw daily network rollup (sums of revenue/patients/admissions/outstanding/
  unbilled), created `WITH NO DATA`, populated by `RefreshNetworkDaily` and (in
  future) the cron worker (GEC-71). A unique index on `metric_date` allows a
  `CONCURRENTLY` refresh later. It backs time-series **charts**, never a headline
  KPI.
- **Partitioning is the documented scale-up path.** At current volume (12
  facilities × daily ≈ a few thousand rows/year) the index is more than enough.
  Declarative **range partitioning of `facility_metrics` by `metric_date`** is the
  scale-up lever, to be enabled only if volume warrants — no partitioning is
  created now.

## Consequences
- Measured on the seeded network (native Postgres 18), the trailing-window query
  is an index scan, not a sequential scan:
  ```
  Index Scan Backward using idx_facility_metrics_facility_date on facility_metrics
    Index Cond: ((facility_id = 'kasoa') AND (metric_date >= '2026-06-24'))
    Buffers: shared hit=2
  Planning Time: 0.016 ms   Execution Time: 0.006 ms
  ```
- KPI parity verified: `kpi.Compute` over the Postgres-loaded series equals
  `kpi.Compute` over the in-memory series — the storage swap changes nothing about
  the numbers.
- The materialized view must be refreshed to stay current; until GEC-71 wires the
  cron, `EnsureSeeded` refreshes it once at first-run seed. A stale view only
  affects charts, never KPIs.
- Deferring partitioning keeps the schema simple; the cost is a future migration
  if volume ever demands it (cheap, and the access pattern already matches a
  time-range partition key).

## Alternatives considered
- **TimescaleDB hypertables / continuous aggregates** — rejected: not available on
  Render's managed Postgres (ADR-0001); would re-introduce the vendor lock OQ-4
  closed.
- **Compute KPIs in SQL (views/MVs as the source of figures)** — rejected:
  violates the "numbers computed in Go" rule and risks SQL/engine divergence in
  front of an executive.
- **Partition `facility_metrics` now** — rejected as premature for 12 facilities;
  documented as the scale-up path instead.
