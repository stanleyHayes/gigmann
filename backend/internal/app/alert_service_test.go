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
	items, next, err := svc.Feed(context.Background(), execPrincipal(), "", 0)
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
	page1, next1, err := svc.Feed(context.Background(), execPrincipal(), "", 2)
	require.NoError(t, err)
	require.Len(t, page1, 2)
	require.NotEmpty(t, next1)

	page2, next2, err := svc.Feed(context.Background(), execPrincipal(), next1, 2)
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

	upd, err := svc.UpdateStatus(ctx, execPrincipal(), "a1", alert.StatusDismissed)
	require.NoError(t, err)
	assert.Equal(t, alert.StatusDismissed, upd.Status)

	_, err = svc.UpdateStatus(ctx, execPrincipal(), "a1", alert.StatusResolved)
	require.ErrorIs(t, err, alert.ErrAlreadyTerminal, "already terminal → 409")

	_, err = svc.UpdateStatus(ctx, execPrincipal(), "ghost", alert.StatusDismissed)
	require.ErrorIs(t, err, ports.ErrAlertNotFound)

	_, err = svc.UpdateStatus(ctx, execPrincipal(), "a1", alert.StatusOpen)
	require.ErrorIs(t, err, app.ErrInvalidAlertStatus, "only dismissed/resolved allowed")
}

func TestAlertFeedScopesManagerToOwnFacility(t *testing.T) {
	t0 := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	repo := memory.NewAlertRepo(
		mkAlert(t, "a-kasoa", "kasoa", severity.Critical, alert.StatusOpen, t0),
		mkAlert(t, "a-nima", "nima", severity.Critical, alert.StatusOpen, t0),
	)
	svc := app.NewAlertService(repo)

	mgr, _, err := svc.Feed(context.Background(), managerPrincipal("kasoa"), "", 0)
	require.NoError(t, err)
	require.Len(t, mgr, 1, "manager sees only their facility's alerts")
	assert.Equal(t, "a-kasoa", mgr[0].ID)

	all, _, err := svc.Feed(context.Background(), execPrincipal(), "", 0)
	require.NoError(t, err)
	assert.Len(t, all, 2, "executive sees the whole network")
}

func TestAlertUpdateStatusRejectsCrossFacility(t *testing.T) {
	t0 := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	repo := memory.NewAlertRepo(mkAlert(t, "a-nima", "nima", severity.Critical, alert.StatusOpen, t0))
	svc := app.NewAlertService(repo)

	_, err := svc.UpdateStatus(context.Background(), managerPrincipal("kasoa"), "a-nima", alert.StatusDismissed)
	require.ErrorIs(t, err, app.ErrForbidden, "a kasoa manager cannot dismiss a nima alert (IDOR)")

	_, err = svc.UpdateStatus(context.Background(), managerPrincipal("nima"), "a-nima", alert.StatusDismissed)
	require.NoError(t, err, "the owning facility's manager may act")
}

func TestAlertFeedCursorResilientToStatusChange(t *testing.T) {
	t0 := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	repo := memory.NewAlertRepo(
		mkAlert(t, "a1", "f1", severity.Critical, alert.StatusOpen, t0),
		mkAlert(t, "a2", "f2", severity.Watch, alert.StatusOpen, t0),
		mkAlert(t, "a3", "f3", severity.Good, alert.StatusOpen, t0),
	)
	svc := app.NewAlertService(repo)

	page1, next, err := svc.Feed(context.Background(), execPrincipal(), "", 1)
	require.NoError(t, err)
	require.Len(t, page1, 1)
	require.NotEmpty(t, next)
	require.Equal(t, "a1", page1[0].ID)

	// The cursor's own alert resolves before page 2 is fetched: paging must still
	// advance past it (not reset to page 1).
	_, err = svc.UpdateStatus(context.Background(), execPrincipal(), page1[0].ID, alert.StatusResolved)
	require.NoError(t, err)

	page2, _, err := svc.Feed(context.Background(), execPrincipal(), next, 5)
	require.NoError(t, err)
	ids := make([]string, 0, len(page2))
	for _, a := range page2 {
		ids = append(ids, a.ID)
	}
	assert.Equal(t, []string{"a2", "a3"}, ids, "advances past the now-resolved cursor; no duplicate of a1")
}

func TestAlertFeedDedupsByFacilityAndType(t *testing.T) {
	t0 := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	repo := memory.NewAlertRepo(
		mkAlertTyped(t, "a-old", "kasoa", "denial", severity.Watch, alert.StatusOpen, t0),
		mkAlertTyped(t, "a-new", "kasoa", "denial", severity.Critical, alert.StatusOpen, t0.Add(time.Hour)),
		mkAlertTyped(t, "a-other", "kasoa", "stockout", severity.Watch, alert.StatusOpen, t0),
	)
	items, _, err := app.NewAlertService(repo).Feed(context.Background(), execPrincipal(), "", 0)
	require.NoError(t, err)
	require.Len(t, items, 2, "the two kasoa/denial alerts collapse to one; kasoa/stockout stays")
	assert.Equal(t, "a-new", items[0].ID, "the most recent of the deduped pair survives, ranked first (critical)")
}
