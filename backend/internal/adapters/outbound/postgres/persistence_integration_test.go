//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/postgres"
	"github.com/xcreativs/gigmann/internal/core/approval"
	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/money"
	"github.com/xcreativs/gigmann/internal/core/payer"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/core/task"
	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/ports"
	"github.com/xcreativs/gigmann/migrations"
)

// testPool is a migrated, shared database for the persistence integration suite.
// One container is started in TestMain; each test truncates first for isolation.
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()
	container, err := tcpostgres.Run(ctx, "pgvector/pgvector:pg16",
		tcpostgres.WithDatabase("gigmann"),
		tcpostgres.WithUsername("gigmann"),
		tcpostgres.WithPassword("gigmann"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "persistence test: start container:", err)
		os.Exit(1)
	}
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintln(os.Stderr, "persistence test: connection string:", err)
		os.Exit(1)
	}
	// Exercise the real migration runner against the embedded migrations.
	if err := postgres.Migrate(ctx, dsn, migrations.Files); err != nil {
		fmt.Fprintln(os.Stderr, "persistence test: migrate:", err)
		os.Exit(1)
	}
	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "persistence test: pool:", err)
		os.Exit(1)
	}

	code := m.Run()

	testPool.Close()
	_ = container.Terminate(ctx)
	os.Exit(code)
}

func truncateAll(ctx context.Context, t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(ctx,
		`TRUNCATE refresh_tokens, credentials, users, approvals, tasks, facilities CASCADE`)
	require.NoError(t, err)
}

func seedFacility(ctx context.Context, t *testing.T, id string) {
	t.Helper()
	mix, err := payer.New(70, 25, 5)
	require.NoError(t, err)
	f, err := facility.New(facility.Params{
		ID: id, Name: id, Region: "Central", Town: "Town", Type: "OPD",
		Beds: 20, Lifecycle: facility.LifecycleActive, Health: severity.Good, PayerMix: mix,
	})
	require.NoError(t, err)
	require.NoError(t, postgres.NewFacilityRepo(testPool).Create(ctx, f))
}

func TestUserRepoIntegration(t *testing.T) {
	ctx := context.Background()
	truncateAll(ctx, t)
	seedFacility(ctx, t, "kasoa")

	repo := postgres.NewUserRepo(testPool)

	ceo, err := user.New(user.User{ID: "u-sammy", Name: "Sammy Adjei", Role: user.RoleExecutive})
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, ports.Account{
		User: ceo, Email: "CEO@Gigmann.Health", PasswordHash: "hash-ceo",
	}))

	mgr, err := user.New(user.User{ID: "u-ama", Name: "Ama Owusu", Role: user.RoleFacilityManager, FacilityID: "kasoa"})
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, ports.Account{
		User: mgr, Email: "ama@gigmann.health", PasswordHash: "hash-ama", MFASecret: "JBSWY3DPEHPK3PXP",
	}))

	// Case-insensitive email lookup, executive has no facility.
	got, err := repo.FindByEmail(ctx, "ceo@gigmann.health")
	require.NoError(t, err)
	assert.Equal(t, "u-sammy", got.User.ID)
	assert.Equal(t, user.RoleExecutive, got.User.Role)
	assert.Empty(t, got.User.FacilityID)
	assert.Equal(t, "ceo@gigmann.health", got.Email)

	// Manager is scoped to a facility; MFA secret round-trips.
	byID, err := repo.FindByID(ctx, "u-ama")
	require.NoError(t, err)
	assert.Equal(t, "kasoa", byID.User.FacilityID)
	assert.Equal(t, "JBSWY3DPEHPK3PXP", byID.MFASecret)
	assert.Equal(t, "hash-ama", byID.PasswordHash)

	// Unknown lookups map to the sentinel error.
	_, err = repo.FindByEmail(ctx, "nobody@gigmann.health")
	require.ErrorIs(t, err, ports.ErrAccountNotFound)
	_, err = repo.FindByID(ctx, "u-nope")
	require.ErrorIs(t, err, ports.ErrAccountNotFound)

	// Save is an upsert: re-save updates the profile in place.
	ceo.Name = "Samuel Adjei"
	require.NoError(t, repo.Save(ctx, ports.Account{User: ceo, Email: "ceo@gigmann.health", PasswordHash: "hash-ceo-2"}))
	updated, err := repo.FindByID(ctx, "u-sammy")
	require.NoError(t, err)
	assert.Equal(t, "Samuel Adjei", updated.User.Name)
	assert.Equal(t, "hash-ceo-2", updated.PasswordHash)
}

func TestRefreshRepoIntegration(t *testing.T) {
	ctx := context.Background()
	truncateAll(ctx, t)

	// A refresh token references a user (FK), so seed one first.
	ceo, err := user.New(user.User{ID: "u-sammy", Name: "Sammy Adjei", Role: user.RoleExecutive})
	require.NoError(t, err)
	require.NoError(t, postgres.NewUserRepo(testPool).Save(ctx, ports.Account{
		User: ceo, Email: "ceo@gigmann.health", PasswordHash: "h",
	}))

	repo := postgres.NewRefreshRepo(testPool)
	principal := auth.Principal{UserID: "u-sammy", Name: "Sammy Adjei", Role: user.RoleExecutive}

	raw, err := repo.Issue(ctx, principal, time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, raw)

	// Consume returns the principal, then the token is single-use.
	got, err := repo.Consume(ctx, raw)
	require.NoError(t, err)
	assert.Equal(t, principal, got)
	_, err = repo.Consume(ctx, raw)
	require.ErrorIs(t, err, ports.ErrInvalidRefreshToken)

	// Revoke invalidates a live token.
	raw2, err := repo.Issue(ctx, principal, time.Hour)
	require.NoError(t, err)
	require.NoError(t, repo.Revoke(ctx, raw2))
	_, err = repo.Consume(ctx, raw2)
	require.ErrorIs(t, err, ports.ErrInvalidRefreshToken)

	// An expired token is rejected (and cleaned up).
	expired, err := repo.Issue(ctx, principal, -time.Minute)
	require.NoError(t, err)
	_, err = repo.Consume(ctx, expired)
	require.ErrorIs(t, err, ports.ErrInvalidRefreshToken)

	// An unknown token is rejected.
	_, err = repo.Consume(ctx, "not-a-real-token")
	require.ErrorIs(t, err, ports.ErrInvalidRefreshToken)

	// Revoking an unknown token is a no-op.
	require.NoError(t, repo.Revoke(ctx, "not-a-real-token"))
}

func TestApprovalRepoIntegration(t *testing.T) {
	ctx := context.Background()
	truncateAll(ctx, t)
	seedFacility(ctx, t, "assin-fosu")

	repo := postgres.NewApprovalRepo(testPool)
	t0 := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)

	withFacility, err := approval.New(approval.Approval{
		ID: "ap-ultrasound", Type: approval.TypeCapital, FacilityID: "assin-fosu",
		Amount: money.FromCedis(85000, 0), Title: "Ultrasound machine", RequestedBy: "Dr. Mensah",
		Status: approval.StatusPending, CreatedAt: t0,
	})
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, withFacility))

	noFacility, err := approval.New(approval.Approval{
		ID: "ap-network", Type: approval.TypeCapital, Title: "Network-wide tooling",
		Amount: money.FromCedis(12000, 50), Status: approval.StatusPending, CreatedAt: t0.Add(time.Hour),
	})
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, noFacility))

	all, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, all, 2)
	// Ordered by created_at: the facility-scoped one was created first.
	assert.Equal(t, "ap-ultrasound", all[0].ID)
	assert.Equal(t, "ap-network", all[1].ID)
	// created_at round-trips exactly (normalised to UTC microseconds on both sides).
	assert.True(t, all[0].CreatedAt.Equal(t0), "created_at must round-trip exactly")
	// Money round-trips exactly (minor units), NULL facility maps to "".
	assert.Equal(t, int64(8500000), all[0].Amount.Pesewas())
	assert.Equal(t, "assin-fosu", all[0].FacilityID)
	assert.Equal(t, int64(1200050), all[1].Amount.Pesewas())
	assert.Empty(t, all[1].FacilityID)

	_, err = repo.Get(ctx, "missing")
	require.ErrorIs(t, err, ports.ErrApprovalNotFound)

	// Decide → Save → Get reflects the decision.
	current, err := repo.Get(ctx, "ap-ultrasound")
	require.NoError(t, err)
	decided, err := current.Decide(true, "Approved for Q3", t0.Add(2*time.Hour))
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, decided))

	reloaded, err := repo.Get(ctx, "ap-ultrasound")
	require.NoError(t, err)
	assert.Equal(t, approval.StatusApproved, reloaded.Status)
	assert.Equal(t, "Approved for Q3", reloaded.DecisionNote)
	assert.WithinDuration(t, t0.Add(2*time.Hour), reloaded.DecidedAt, time.Second)
}

func TestTaskRepoIntegration(t *testing.T) {
	ctx := context.Background()
	truncateAll(ctx, t)
	seedFacility(ctx, t, "kasoa")

	repo := postgres.NewTaskRepo(testPool)
	t0 := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	due := time.Date(2026, 6, 5, 17, 0, 0, 0, time.UTC)

	scoped, err := task.New(task.Task{
		ID: "task-kasoa-denials", Title: "Review NHIS denial spike at Kasoa", Detail: "Denial rate at 19%.",
		FacilityID: "kasoa", Priority: task.PriorityHigh, Status: task.StatusInProgress,
		Source: task.SourceAlert, DueDate: due, CreatedAt: t0,
	})
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, scoped))

	board, err := task.New(task.Task{
		ID: "task-board-deck", Title: "Finalise Q3 board deck",
		Priority: task.PriorityMedium, Status: task.StatusTodo, Source: task.SourceManual,
		CreatedAt: t0.Add(time.Hour),
	})
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, board))

	all, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, all, 2)
	assert.Equal(t, "task-kasoa-denials", all[0].ID)
	assert.Equal(t, "kasoa", all[0].FacilityID)
	assert.WithinDuration(t, due, all[0].DueDate, time.Second)
	// NULL facility + NULL due date map back to zero values.
	assert.Empty(t, all[1].FacilityID)
	assert.True(t, all[1].DueDate.IsZero())

	_, err = repo.Get(ctx, "missing")
	require.ErrorIs(t, err, ports.ErrTaskNotFound)

	// Status update round-trips.
	current, err := repo.Get(ctx, "task-kasoa-denials")
	require.NoError(t, err)
	current.Status = task.StatusDone
	require.NoError(t, repo.Save(ctx, current))
	reloaded, err := repo.Get(ctx, "task-kasoa-denials")
	require.NoError(t, err)
	assert.Equal(t, task.StatusDone, reloaded.Status)
}

func TestEnsureSeededIdempotencyIntegration(t *testing.T) {
	ctx := context.Background()
	truncateAll(ctx, t)

	mix, err := payer.New(70, 25, 5)
	require.NoError(t, err)
	fac, err := facility.New(facility.Params{
		ID: "kasoa", Name: "Kasoa", Region: "Central", Lifecycle: facility.LifecycleActive,
		Health: severity.Good, PayerMix: mix,
	})
	require.NoError(t, err)
	t0 := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	appr, err := approval.New(approval.Approval{
		ID: "ap-x", Type: approval.TypeCapital, FacilityID: "kasoa", Amount: money.FromCedis(1000, 0),
		Title: "Thing", Status: approval.StatusPending, CreatedAt: t0,
	})
	require.NoError(t, err)
	tsk, err := task.New(task.Task{
		ID: "task-x", Title: "Do", Priority: task.PriorityLow, Status: task.StatusTodo,
		Source: task.SourceManual, CreatedAt: t0,
	})
	require.NoError(t, err)
	ceo, err := user.New(user.User{ID: "u-sammy", Name: "Sammy Adjei", Role: user.RoleExecutive})
	require.NoError(t, err)

	facs := []facility.Facility{fac}
	apprs := []approval.Approval{appr}
	tasks := []task.Task{tsk}
	accounts := []ports.Account{{User: ceo, Email: "ceo@gigmann.health", PasswordHash: "h"}}

	// First run seeds an empty database.
	seeded, err := postgres.EnsureSeeded(ctx, testPool, facs, apprs, tasks, accounts)
	require.NoError(t, err)
	assert.True(t, seeded, "first run must seed")

	// Mutate persisted data (decide the seeded approval).
	apprRepo := postgres.NewApprovalRepo(testPool)
	cur, err := apprRepo.Get(ctx, "ap-x")
	require.NoError(t, err)
	decided, err := cur.Decide(true, "approved", t0.Add(time.Hour))
	require.NoError(t, err)
	require.NoError(t, apprRepo.Save(ctx, decided))

	// A subsequent run (restart) must NOT re-seed, preserving the mutation.
	seeded2, err := postgres.EnsureSeeded(ctx, testPool, facs, apprs, tasks, accounts)
	require.NoError(t, err)
	assert.False(t, seeded2, "restart must not re-seed a populated database")

	got, err := apprRepo.Get(ctx, "ap-x")
	require.NoError(t, err)
	assert.Equal(t, approval.StatusApproved, got.Status, "restart must preserve the decided approval")

	all, err := postgres.NewFacilityRepo(testPool).List(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1, "no duplicate facilities from a second EnsureSeeded")
}
