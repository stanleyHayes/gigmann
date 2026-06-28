// Package webpush is the outbound adapter that delivers encrypted Web Push
// payloads (RFC 8291) signed with VAPID (RFC 8292). It implements
// ports.PushSender. When VAPID keys are not configured the sender is disabled
// and Send is a no-op, so push degrades to off without keys (no panics, no
// errors) — mirroring how the app treats the Anthropic/Voyage/Sentry keys.
package webpush

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/xcreativs/gigmann/internal/ports"
)

const pushTTL = 24 * time.Hour

// Sender delivers Web Push notifications. The zero-config (empty keys) Sender is
// valid and disabled.
type Sender struct {
	publicKey  string
	privateKey string
	subject    string // VAPID "sub" claim, e.g. mailto:ops@gigmann.health
	client     *http.Client
}

// New constructs a Sender. With an empty public or private key it is disabled
// (Enabled reports false and Send is a no-op). subject falls back to a mailto.
func New(publicKey, privateKey, subject string) *Sender {
	if subject == "" {
		subject = "mailto:ops@gigmann.health"
	}
	return &Sender{
		publicKey:  publicKey,
		privateKey: privateKey,
		subject:    subject,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

var _ ports.PushSender = (*Sender)(nil)

// Enabled reports whether VAPID keys are configured.
func (s *Sender) Enabled() bool { return s.publicKey != "" && s.privateKey != "" }

// PublicKey returns the VAPID application server public key (base64url) clients
// need to subscribe, or "" when push is not configured.
func (s *Sender) PublicKey() string { return s.publicKey }

// Send delivers one encrypted payload to one subscription. It is a no-op when
// disabled. A 404/410 from the push service means the subscription expired; the
// caller may prune it — surfaced as ErrGone.
func (s *Sender) Send(ctx context.Context, sub ports.PushSubscription, payload []byte) error {
	if !s.Enabled() {
		return nil
	}
	resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys:     webpush.Keys{P256dh: sub.P256dh, Auth: sub.Auth},
	}, &webpush.Options{
		Subscriber:      s.subject,
		VAPIDPublicKey:  s.publicKey,
		VAPIDPrivateKey: s.privateKey,
		TTL:             int(pushTTL.Seconds()),
		HTTPClient:      s.client,
	})
	if err != nil {
		return fmt.Errorf("webpush: send: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		return ErrGone
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webpush: push service returned %d", resp.StatusCode)
	}
	return nil
}

// ErrGone indicates the push subscription is no longer valid and should be pruned.
var ErrGone = errors.New("webpush: subscription gone")
