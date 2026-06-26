-- 000003_metrics_views.up.sql — charting aggregate for facility_metrics (GEC-12).
-- KPI numbers are always computed by the Go signal engine (kpi.Compute); this
-- materialized view is a RAW daily network rollup for time-series charts and the
-- scale-up path only — never a source of KPI truth. Created WITH NO DATA; the
-- cron worker (GEC-71) / RefreshNetworkDaily populates it. The unique index lets
-- a future refresh run CONCURRENTLY. sum() over bigint yields numeric, so each
-- money/count column is cast back to bigint for clean integer scans.
CREATE MATERIALIZED VIEW network_daily_metrics AS
SELECT metric_date,
       sum(revenue)::bigint          AS revenue,
       sum(patients_seen)::bigint    AS patients_seen,
       sum(admissions)::bigint       AS admissions,
       sum(nhis_outstanding)::bigint AS nhis_outstanding,
       sum(unbilled_amount)::bigint  AS unbilled_amount
FROM facility_metrics
GROUP BY metric_date
WITH NO DATA;

CREATE UNIQUE INDEX idx_network_daily_metrics_date ON network_daily_metrics (metric_date);
