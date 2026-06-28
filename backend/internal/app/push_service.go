package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/xcreativs/gigmann/internal/core/alert"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/ports"
)

// PushService manages Web Push subscriptions and delivers *critical-only* alert
// notifications (spec §5.11 "quiet by default"): a notification is sent only for
// a severity=critical alert, only to users who explicitly opted in by
// subscribing, and only once per (subscription, alert) — so the channel stays
// trusted. It is a no-op when the sender has no VAPID keys.
type PushService struct {
	store  ports.PushSubscriptionStore
	sender ports.PushSender
	alerts ports.AlertRepository

	mu   sync.Mutex
	sent map[string]bool // "endpoint\x00alertID" already delivered (dedup)
}

// NewPushService wires the subscription store, the (possibly disabled) sender,
// and the alert source used by the critical sweep.
func NewPushService(store ports.PushSubscriptionStore, sender ports.PushSender, alerts ports.AlertRepository) *PushService {
	return &PushService{store: store, sender: sender, alerts: alerts, sent: map[string]bool{}}
}

// Enabled reports whether push is configured (VAPID keys present).
func (s *PushService) Enabled() bool { return s.sender.Enabled() }

// PublicKey returns the VAPID public key clients need to subscribe ("" if off).
func (s *PushService) PublicKey() string { return s.sender.PublicKey() }

// Subscribe records a browser subscription for the user (idempotent per
// endpoint) and immediately delivers any currently-open critical alerts to it,
// so a freshly-enabled device is caught up. Inputs are validated.
func (s *PushService) Subscribe(ctx context.Context, userID string, sub ports.PushSubscription) error {
	if err := validateSubscription(sub); err != nil {
		return err
	}
	if err := s.store.Save(ctx, userID, sub); err != nil {
		return fmt.Errorf("push: save subscription: %w", err)
	}
	// Best-effort catch-up for the new device; never blocks subscription.
	_, _ = s.Sweep(ctx)
	return nil
}

// Unsubscribe removes a subscription for the user (idempotent).
func (s *PushService) Unsubscribe(ctx context.Context, userID, endpoint string) error {
	if strings.TrimSpace(endpoint) == "" {
		return ErrInvalidPushSubscription
	}
	if err := s.store.Delete(ctx, userID, endpoint); err != nil {
		return fmt.Errorf("push: delete subscription: %w", err)
	}
	return nil
}

// Notify implements ports.Notifier so the service can hang off the same
// brief-refresh signal as the realtime hub: each refresh triggers a sweep.
func (s *PushService) Notify(string) {
	if !s.sender.Enabled() {
		return
	}
	go func() { _, _ = s.Sweep(context.Background()) }()
}

// Sweep delivers every currently-open critical alert to every subscription that
// has not already received it, and returns the number of notifications sent. It
// is a no-op when push is disabled.
func (s *PushService) Sweep(ctx context.Context) (int, error) {
	if !s.sender.Enabled() {
		return 0, nil
	}
	alerts, err := s.alerts.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("push: list alerts: %w", err)
	}
	byUser, err := s.store.All(ctx)
	if err != nil {
		return 0, fmt.Errorf("push: list subscriptions: %w", err)
	}

	sent := 0
	for _, a := range alerts {
		if a.Severity == severity.Critical && a.Status == alert.StatusOpen {
			sent += s.deliver(ctx, a, byUser)
		}
	}
	return sent, nil
}

// deliver sends one critical alert to every subscription that has not already
// received it, returning how many notifications were delivered.
func (s *PushService) deliver(ctx context.Context, a alert.Alert, byUser map[string][]ports.PushSubscription) int {
	payload := s.payloadFor(a)
	sent := 0
	for _, subs := range byUser {
		for _, sub := range subs {
			if s.markSent(sub.Endpoint, a.ID) {
				continue // already delivered to this device
			}
			if err := s.sender.Send(ctx, sub, payload); err == nil {
				sent++
			}
		}
	}
	return sent
}

// markSent records a (subscription, alert) pair as delivered and reports whether
// it had already been recorded.
func (s *PushService) markSent(endpoint, alertID string) bool {
	key := endpoint + "\x00" + alertID
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sent[key] {
		return true
	}
	s.sent[key] = true
	return false
}

func (s *PushService) payloadFor(a alert.Alert) []byte {
	// The browser SW reads these fields; figures come straight from the
	// deterministic alert (the AI never invents notification numbers).
	b, err := json.Marshal(map[string]string{
		"title":      a.Title,
		"body":       a.Detail,
		"alertId":    a.ID,
		"facilityId": a.FacilityID,
		"url":        "/alerts",
	})
	if err != nil {
		return []byte(`{}`)
	}
	return b
}

// ErrInvalidPushSubscription is returned when a subscription is missing required fields.
var ErrInvalidPushSubscription = errors.New("push: invalid subscription")

func validateSubscription(sub ports.PushSubscription) error {
	if strings.TrimSpace(sub.Endpoint) == "" ||
		strings.TrimSpace(sub.P256dh) == "" ||
		strings.TrimSpace(sub.Auth) == "" {
		return ErrInvalidPushSubscription
	}
	// Only accept https push endpoints (web push services are always TLS).
	if !strings.HasPrefix(sub.Endpoint, "https://") {
		return ErrInvalidPushSubscription
	}
	return nil
}
