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

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: worker <migrate|refresh-views>")
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("worker: config: %v", err)
	}
	if cfg.DatabaseURL == "" {
		log.Fatal("worker: DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	if err := run(ctx, os.Args[1], cfg.DatabaseURL); err != nil {
		log.Fatalf("worker: %s: %v", os.Args[1], err)
	}
	log.Printf("worker: %s done", os.Args[1])
}

func run(ctx context.Context, job, dsn string) error {
	// Schema is reconciled first (idempotent, advisory-locked) so the job is safe
	// even on a fresh database.
	if err := postgres.Migrate(ctx, dsn, migrations.Files); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	switch job {
	case "migrate":
		return nil
	case "refresh-views":
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
