-- name: ListNetworkMetrics :many
SELECT facility_id, metric_date, revenue, cash_revenue, momo_revenue,
       patients_seen, admissions, occupancy_rate, avg_wait_minutes,
       nhis_claims_submitted, nhis_claims_paid, nhis_claims_denied,
       nhis_outstanding, unbilled_amount
FROM facility_metrics
ORDER BY facility_id, metric_date;

-- name: ListFacilityMetricsSince :many
SELECT facility_id, metric_date, revenue, cash_revenue, momo_revenue,
       patients_seen, admissions, occupancy_rate, avg_wait_minutes,
       nhis_claims_submitted, nhis_claims_paid, nhis_claims_denied,
       nhis_outstanding, unbilled_amount
FROM facility_metrics
WHERE facility_id = $1 AND metric_date >= $2
ORDER BY metric_date;

-- name: InsertFacilityMetric :exec
INSERT INTO facility_metrics (
    facility_id, metric_date, revenue, cash_revenue, momo_revenue,
    patients_seen, admissions, occupancy_rate, avg_wait_minutes,
    nhis_claims_submitted, nhis_claims_paid, nhis_claims_denied,
    nhis_outstanding, unbilled_amount
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
ON CONFLICT (facility_id, metric_date) DO UPDATE SET
    revenue = EXCLUDED.revenue,
    cash_revenue = EXCLUDED.cash_revenue,
    momo_revenue = EXCLUDED.momo_revenue,
    patients_seen = EXCLUDED.patients_seen,
    admissions = EXCLUDED.admissions,
    occupancy_rate = EXCLUDED.occupancy_rate,
    avg_wait_minutes = EXCLUDED.avg_wait_minutes,
    nhis_claims_submitted = EXCLUDED.nhis_claims_submitted,
    nhis_claims_paid = EXCLUDED.nhis_claims_paid,
    nhis_claims_denied = EXCLUDED.nhis_claims_denied,
    nhis_outstanding = EXCLUDED.nhis_outstanding,
    unbilled_amount = EXCLUDED.unbilled_amount;
