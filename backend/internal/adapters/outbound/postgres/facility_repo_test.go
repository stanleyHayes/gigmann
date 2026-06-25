//go:build integration

package postgres_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/postgres"
	"github.com/xcreativs/gigmann/internal/core/severity"
)

func TestFacilityRepoListIntegration(t *testing.T) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "pgvector/pgvector:pg16",
		tcpostgres.WithDatabase("gigmann"),
		tcpostgres.WithUsername("gigmann"),
		tcpostgres.WithPassword("gigmann"),
		tcpostgres.BasicWaitStrategies(),
	)
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Simple protocol lets us exec the multi-statement schema in one call.
	cfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)
	defer pool.Close()

	_, err = pool.Exec(ctx, readSchema(t))
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO facilities (id, name, region, town, type, beds, lifecycle, health,
			manager_name, payer_nhis, payer_cash_momo, payer_private, latitude, longitude)
		VALUES ('kasoa','Kasoa Polyclinic','Central','Kasoa','OPD',40,'active','watch',
			'Ama Owusu',70,25,5,5.53,-0.42)`)
	require.NoError(t, err)

	repo := postgres.NewFacilityRepo(pool)
	got, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "Kasoa Polyclinic", got[0].Name)
	require.Equal(t, 70, got[0].PayerMix.NHIS)
	require.Equal(t, severity.Watch, got[0].Health)
}

func readSchema(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	// backend/ is four levels up from internal/adapters/outbound/postgres/.
	path := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "migrations", "000001_init.up.sql")
	b, err := os.ReadFile(filepath.Clean(path))
	require.NoError(t, err)
	return string(b)
}
