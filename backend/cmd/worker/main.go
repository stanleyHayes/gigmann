// Command worker runs scheduled maintenance jobs (GEC-71): refreshing the
// materialized view and applying migrations. It is a thin composition root over
// the same outbound adapters as the API, run by the Render cron schedule.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/postgres"
	"github.com/xcreativs/gigmann/internal/config"
	"github.com/xcreativs/gigmann/migrations"
)

const jobTimeout = 2 * time.Minute

// Valid job names. main validates os.Args[1] against these before using it, so
// only a known literal is ever interpolated into logs (avoids log injection, gosec G706).
const (
	jobMigrate      = "migrate"
	jobRefreshViews = "refresh-views"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: worker <migrate|refresh-views>")
	}
	var job string
	switch os.Args[1] {
	case jobMigrate:
		job = jobMigrate
	case jobRefreshViews:
		job = jobRefreshViews
	default:
		log.Fatal("usage: worker <migrate|refresh-views>")
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("worker: config: %v", err)
	}
	if cfg.DatabaseURL == "" {
		log.Fatal("worker: DATABASE_URL is required")
	}

	if err := execute(job, cfg.DatabaseURL); err != nil {
		log.Fatalf("worker: %s: %v", job, err)
	}
	log.Printf("worker: %s done", job)
}

// execute scopes the job timeout so the deferred cancel runs before main exits.
func execute(job, dsn string) error {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()
	return run(ctx, job, dsn)
}

func run(ctx context.Context, job, dsn string) error {
	// Schema is reconciled first (idempotent, advisory-locked) so the job is safe
	// even on a fresh database.
	if err := postgres.Migrate(ctx, dsn, migrations.Files); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	switch job {
	case jobMigrate:
		return nil
	case jobRefreshViews:
		pool, err := postgres.Connect(ctx, dsn)
		if err != nil {
			return err
		}
		defer pool.Close()
		return postgres.NewMetricsRepo(pool).RefreshNetworkDaily(ctx)
	default:
		return fmt.Errorf("unknown job %q (want migrate|refresh-views)", job)
	}
}
