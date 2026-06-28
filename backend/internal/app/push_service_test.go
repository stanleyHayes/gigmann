package app_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/alert"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/ports"
)

// fakeSender is a ports.PushSender that records deliveries and can be toggled on/off.
type fakeSender struct {
	enabled bool
	mu      sync.Mutex
	sent    []ports.PushSubscription
}

func (f *fakeSender) Enabled() bool     { return f.enabled }
func (f *fakeSender) PublicKey() string { return "test-public-key" }
func (f *fakeSender) Send(_ context.Context, sub ports.PushSubscription, _ []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, sub)
	return nil
}
func (f *fakeSender) count() int { f.mu.Lock(); defer f.mu.Unlock(); return len(f.sent) }

// fakeAlertRepo is a ports.AlertRepository returning a fixed list.
type fakeAlertRepo struct{ alerts []alert.Alert }

func (r *fakeAlertRepo) List(context.Context) ([]alert.Alert, error) { return r.alerts, nil }
func (r *fakeAlertRepo) Get(context.Context, string) (alert.Alert, error) {
	return alert.Alert{}, ports.ErrAlertNotFound
}
func (r *fakeAlertRepo) Save(context.Context, alert.Alert) error { return nil }

func validSub() ports.PushSubscription {
	return ports.PushSubscription{Endpoint: "https://push.example.com/abc", P256dh: "key", Auth: "secret"}
}

func criticalAndWatch() []alert.Alert {
	return []alert.Alert{
		{ID: "a-crit", FacilityID: "kasoa", Type: "stockout", Severity: severity.Critical, Title: "Stock-out imminent", Detail: "2 days", Status: alert.StatusOpen},
		{ID: "a-watch", FacilityID: "nima", Type: "waits", Severity: severity.Watch, Title: "Waits up", Detail: "x", Status: alert.StatusOpen},
		{ID: "a-resolved", FacilityID: "kasoa", Type: "stockout", Severity: severity.Critical, Title: "old", Detail: "x", Status: alert.StatusResolved},
	}
}

func TestPushService_DisabledIsNoop(t *testing.T) {
	sender := &fakeSender{enabled: false}
	svc := app.NewPushService(memory.NewPushRepo(), sender, &fakeAlertRepo{alerts: criticalAndWatch()})

	assert.False(t, svc.Enabled())
	require.NoError(t, svc.Subscribe(context.Background(), "u1", validSub()))
	n, err := svc.Sweep(context.Background())
	require.NoError(t, err)
	assert.Zero(t, n)
	assert.Zero(t, sender.count())
}

func TestPushService_SubscribeDeliversCriticalOnly(t *testing.T) {
	sender := &fakeSender{enabled: true}
	svc := app.NewPushService(memory.NewPushRepo(), sender, &fakeAlertRepo{alerts: criticalAndWatch()})

	require.NoError(t, svc.Subscribe(context.Background(), "u1", validSub()))
	// Subscribe runs an immediate catch-up sweep: exactly the one open critical alert.
	assert.Equal(t, 1, sender.count())

	// A second sweep is deduped per (endpoint, alert): no re-send.
	n, err := svc.Sweep(context.Background())
	require.NoError(t, err)
	assert.Zero(t, n)
	assert.Equal(t, 1, sender.count())
}

func TestPushService_RejectsInvalidSubscription(t *testing.T) {
	svc := app.NewPushService(memory.NewPushRepo(), &fakeSender{enabled: true}, &fakeAlertRepo{})
	for _, bad := range []ports.PushSubscription{
		{Endpoint: "", P256dh: "k", Auth: "a"},
		{Endpoint: "https://x", P256dh: "", Auth: "a"},
		{Endpoint: "http://insecure", P256dh: "k", Auth: "a"}, // non-https rejected
	} {
		err := svc.Subscribe(context.Background(), "u1", bad)
		assert.ErrorIs(t, err, app.ErrInvalidPushSubscription, "want invalid for %+v", bad)
	}
}

func TestPushService_Unsubscribe(t *testing.T) {
	store := memory.NewPushRepo()
	svc := app.NewPushService(store, &fakeSender{enabled: true}, &fakeAlertRepo{})
	require.NoError(t, svc.Subscribe(context.Background(), "u1", validSub()))

	subs, _ := store.ListByUser(context.Background(), "u1")
	require.Len(t, subs, 1)

	require.NoError(t, svc.Unsubscribe(context.Background(), "u1", validSub().Endpoint))
	subs, _ = store.ListByUser(context.Background(), "u1")
	assert.Empty(t, subs)

	assert.ErrorIs(t, svc.Unsubscribe(context.Background(), "u1", ""), app.ErrInvalidPushSubscription)
}

func TestPushService_NotifyDisabledReturns(t *testing.T) {
	svc := app.NewPushService(memory.NewPushRepo(), &fakeSender{enabled: false}, &fakeAlertRepo{})
	svc.Notify("brief.refreshed") // disabled: returns without spawning work
	assert.Equal(t, "test-public-key", svc.PublicKey())
}

func TestFanoutNotifier(t *testing.T) {
	a, b := &countNotifier{}, &countNotifier{}
	app.FanoutNotifier(a, b).Notify("brief.refreshed")
	assert.Equal(t, 1, a.n)
	assert.Equal(t, 1, b.n)
}

type countNotifier struct{ n int }

func (c *countNotifier) Notify(string) { c.n++ }
