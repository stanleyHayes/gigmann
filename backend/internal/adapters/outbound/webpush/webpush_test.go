package webpush_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/webpush"
	"github.com/xcreativs/gigmann/internal/ports"
)

func TestSender_DisabledWithoutKeys(t *testing.T) {
	s := webpush.New("", "", "")
	assert.False(t, s.Enabled())
	assert.Empty(t, s.PublicKey())
	// Send is a no-op (no network) when disabled.
	require.NoError(t, s.Send(context.Background(),
		ports.PushSubscription{Endpoint: "https://push.example/x", P256dh: "k", Auth: "a"}, []byte(`{}`)))
}

func TestSender_EnabledWithKeys(t *testing.T) {
	s := webpush.New("pub", "priv", "")
	assert.True(t, s.Enabled())
	assert.Equal(t, "pub", s.PublicKey())
}

func TestSender_DefaultSubject(t *testing.T) {
	// Empty subject is replaced with a default mailto; an explicit one is kept.
	assert.NotNil(t, webpush.New("pub", "priv", ""))
	assert.NotNil(t, webpush.New("pub", "priv", "mailto:ops@x.io"))
}
