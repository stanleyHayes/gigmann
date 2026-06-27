package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/alert"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/ports"
)

func mkAlert(t *testing.T, id, fid string, sev severity.Severity, status alert.Status, created time.Time) alert.Alert {
	t.Helper()
	a, err := alert.New(alert.Alert{
		ID: id, FacilityID: fid, Type: "signal", Severity: sev, Title: id, Status: status, CreatedAt: created,
	})
	require.NoError(t, err)
	return a
}

func mkAlertTyped(t *testing.T, id, fid, typ string, sev severity.Severity, status alert.Status, created time.Time) alert.Alert {
	t.Helper()
	a, err := alert.New(alert.Alert{ID: id, FacilityID: fid, Type: typ, Severity: sev, Title: id, Status: status, CreatedAt: created})
	require.NoError(t, err)
	return a
}

func TestAlertFeedRanksOpenWorstFirst(t *testing.T) {
	t0 := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	repo := memory.NewAlertRepo(
		mkAlert(t, "a-watch", "kasoa", severity.Watch, alert.StatusOpen, t0),
		mkAlert(t, "a-crit", "tafo", severity.Critical, alert.StatusOpen, t0),
		mkAlert(t, "a-resolved", "nima", severity.Critical, alert.StatusResolved, t0),
		mkAlert(t, "a-good", "adansi", severity.Good, alert.StatusOpen, t0),
	)
	svc := app.NewAlertService(repo)
	items, next, err := svc.Feed(context.Background(), "", 0)
	require.NoError(t, err)
	require.Len(t, items, 3, "resolved alert drops off the feed")
	assert.Equal(t, "a-crit", items[0].ID, "worst first")
	assert.Equal(t, "a-watch", items[1].ID)
	assert.Equal(t, "a-good", items[2].ID)
	assert.Empty(t, next, "single page")
}

func TestAlertFeedPaginatesByCursor(t *testing.T) {
	t0 := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	repo := memory.NewAlertRepo(
		mkAlert(t, "a1", "f1", severity.Critical, alert.StatusOpen, t0),
		mkAlert(t, "a2", "f2", severity.Watch, alert.StatusOpen, t0),
		mkAlert(t, "a3", "f3", severity.Good, alert.StatusOpen, t0),
	)
	svc := app.NewAlertService(repo)
	page1, next1, err := svc.Feed(context.Background(), "", 2)
	require.NoError(t, err)
	require.Len(t, page1, 2)
	require.NotEmpty(t, next1)

	page2, next2, err := svc.Feed(context.Background(), next1, 2)
	require.NoError(t, err)
	require.Len(t, page2, 1)
	assert.Empty(t, next2)
	assert.NotEqual(t, page1[1].ID, page2[0].ID, "no overlap across pages")
}

func TestAlertUpdateStatusTransitions(t *testing.T) {
	t0 := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	repo := memory.NewAlertRepo(mkAlert(t, "a1", "kasoa", severity.Critical, alert.StatusOpen, t0))
	svc := app.NewAlertService(repo)
	ctx := context.Background()

	upd, err := svc.UpdateStatus(ctx, "a1", alert.StatusDismissed)
	require.NoError(t, err)
	assert.Equal(t, alert.StatusDismissed, upd.Status)

	_, err = svc.UpdateStatus(ctx, "a1", alert.StatusResolved)
	require.ErrorIs(t, err, alert.ErrAlreadyTerminal, "already terminal → 409")

	_, err = svc.UpdateStatus(ctx, "ghost", alert.StatusDismissed)
	require.ErrorIs(t, err, ports.ErrAlertNotFound)

	_, err = svc.UpdateStatus(ctx, "a1", alert.StatusOpen)
	require.ErrorIs(t, err, app.ErrInvalidAlertStatus, "only dismissed/resolved allowed")
}

func TestAlertFeedDedupsByFacilityAndType(t *testing.T) {
	t0 := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	repo := memory.NewAlertRepo(
		mkAlertTyped(t, "a-old", "kasoa", "denial", severity.Watch, alert.StatusOpen, t0),
		mkAlertTyped(t, "a-new", "kasoa", "denial", severity.Critical, alert.StatusOpen, t0.Add(time.Hour)),
		mkAlertTyped(t, "a-other", "kasoa", "stockout", severity.Watch, alert.StatusOpen, t0),
	)
	items, _, err := app.NewAlertService(repo).Feed(context.Background(), "", 0)
	require.NoError(t, err)
	require.Len(t, items, 2, "the two kasoa/denial alerts collapse to one; kasoa/stockout stays")
	assert.Equal(t, "a-new", items[0].ID, "the most recent of the deduped pair survives, ranked first (critical)")
}
