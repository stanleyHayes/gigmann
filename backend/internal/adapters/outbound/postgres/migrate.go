package postgres

import (
	"context"
	"fmt"
	"io/fs"
	"sort"

	"github.com/jackc/pgx/v5"
)

// migrateAdvisoryLock is a fixed key all instances use to serialize migrations.
const migrateAdvisoryLock int64 = 8163264

// Migrate applies, in version (filename) order, every *.up.sql migration in files
// that has not yet been recorded in schema_migrations. Each migration runs in its
// own transaction. It uses the simple query protocol so a multi-statement migration
// file executes as a single command, and it is idempotent: already-applied
// migrations are skipped, so it is safe to call on every startup.
func Migrate(ctx context.Context, dsn string, files fs.FS) error {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("postgres: parse dsn for migrate: %w", err)
	}
	cfg.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("postgres: migrate connect: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	// Serialize concurrent starters (rolling deploy / multiple replicas): only one
	// process applies migrations at a time; the others block here, then find every
	// migration already applied. The lock auto-releases on unlock or conn close.
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, migrateAdvisoryLock); err != nil {
		return fmt.Errorf("postgres: acquire migration lock: %w", err)
	}
	defer func() { _, _ = conn.Exec(ctx, `SELECT pg_advisory_unlock($1)`, migrateAdvisoryLock) }()

	if _, err := conn.Exec(ctx,
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version    text PRIMARY KEY,
			applied_at timestamptz NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("postgres: ensure schema_migrations: %w", err)
	}

	names, err := fs.Glob(files, "*.up.sql")
	if err != nil {
		return fmt.Errorf("postgres: glob migrations: %w", err)
	}
	sort.Strings(names)

	for _, name := range names {
		var applied bool
		if err := conn.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, name,
		).Scan(&applied); err != nil {
			return fmt.Errorf("postgres: check migration %q: %w", name, err)
		}
		if applied {
			continue
		}

		body, err := fs.ReadFile(files, name)
		if err != nil {
			return fmt.Errorf("postgres: read migration %q: %w", name, err)
		}

		tx, err := conn.Begin(ctx)
		if err != nil {
			return fmt.Errorf("postgres: begin migration %q: %w", name, err)
		}
		if _, err := tx.Exec(ctx, string(body)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("postgres: apply migration %q: %w", name, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("postgres: record migration %q: %w", name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("postgres: commit migration %q: %w", name, err)
		}
	}
	return nil
}
