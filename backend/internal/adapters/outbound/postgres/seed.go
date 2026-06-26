package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/postgres/sqlcgen"
	"github.com/xcreativs/gigmann/internal/core/approval"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/metric"
	"github.com/xcreativs/gigmann/internal/core/task"
	"github.com/xcreativs/gigmann/internal/ports"
)

// EnsureSeeded populates an empty database with the given network and accounts,
// and is a no-op when the database already holds facilities. It reports whether
// it seeded. Because the seed runs in a single transaction (see Seed), a database
// is never left half-populated, so the "facilities is empty" guard reliably means
// "no complete seed has run" — a restart therefore preserves persisted changes
// (decided approvals, completed tasks).
func EnsureSeeded(
	ctx context.Context, pool *pgxpool.Pool,
	facs []facility.Facility, metrics []metric.FacilityMetric,
	apprs []approval.Approval, tasks []task.Task, accounts []ports.Account,
) (bool, error) {
	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM facilities`).Scan(&n); err != nil {
		return false, fmt.Errorf("postgres: count facilities: %w", err)
	}
	if n > 0 {
		return false, nil
	}
	if err := Seed(ctx, pool, facs, metrics, apprs, tasks, accounts); err != nil {
		return false, err
	}
	// Populate the charting materialized view from the freshly-seeded data.
	if err := NewMetricsRepo(pool).RefreshNetworkDaily(ctx); err != nil {
		return false, err
	}
	return true, nil
}

// Seed inserts the network and accounts in a SINGLE transaction (all-or-nothing).
// FK order is honoured: facilities first (approvals/tasks/users reference them),
// then approvals, tasks, and finally each account's user+credentials. A failure
// rolls the whole thing back, leaving the database empty for the next attempt.
func Seed(
	ctx context.Context, pool *pgxpool.Pool,
	facs []facility.Facility, metrics []metric.FacilityMetric,
	apprs []approval.Approval, tasks []task.Task, accounts []ports.Account,
) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres: begin seed: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op once committed

	q := sqlcgen.New(tx)
	for _, f := range facs {
		if err := q.CreateFacility(ctx, facilityParams(f)); err != nil {
			return fmt.Errorf("postgres: seed facility %q: %w", f.ID, err)
		}
	}
	for _, m := range metrics {
		if err := q.InsertFacilityMetric(ctx, metricParams(m)); err != nil {
			return fmt.Errorf("postgres: seed metric %s/%s: %w", m.FacilityID, m.Date.Format("2006-01-02"), err)
		}
	}
	for _, a := range apprs {
		if err := q.UpsertApproval(ctx, approvalParams(a)); err != nil {
			return fmt.Errorf("postgres: seed approval %q: %w", a.ID, err)
		}
	}
	for _, t := range tasks {
		if err := q.UpsertTask(ctx, taskParams(t)); err != nil {
			return fmt.Errorf("postgres: seed task %q: %w", t.ID, err)
		}
	}
	for _, acct := range accounts {
		if err := saveAccountTx(ctx, q, acct); err != nil {
			return fmt.Errorf("postgres: seed account %q: %w", acct.User.ID, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres: commit seed: %w", err)
	}
	return nil
}
