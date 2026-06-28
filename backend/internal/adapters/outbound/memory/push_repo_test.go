package memory_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/ports"
)

func sub(endpoint string) ports.PushSubscription {
	return ports.PushSubscription{Endpoint: endpoint, P256dh: "k", Auth: "a"}
}

func TestPushRepo_SaveDedupAndList(t *testing.T) {
	r := memory.NewPushRepo()
	ctx := context.Background()

	require.NoError(t, r.Save(ctx, "u1", sub("https://e/1")))
	require.NoError(t, r.Save(ctx, "u1", sub("https://e/1"))) // same endpoint -> deduped
	require.NoError(t, r.Save(ctx, "u1", sub("https://e/2")))
	require.NoError(t, r.Save(ctx, "u2", sub("https://e/3")))

	u1, err := r.ListByUser(ctx, "u1")
	require.NoError(t, err)
	assert.Len(t, u1, 2)

	all, err := r.All(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 2)
	assert.Len(t, all["u1"], 2)
	assert.Len(t, all["u2"], 1)
}

func TestPushRepo_Delete(t *testing.T) {
	r := memory.NewPushRepo()
	ctx := context.Background()
	require.NoError(t, r.Save(ctx, "u1", sub("https://e/1")))
	require.NoError(t, r.Save(ctx, "u1", sub("https://e/2")))

	require.NoError(t, r.Delete(ctx, "u1", "https://e/1"))
	u1, _ := r.ListByUser(ctx, "u1")
	assert.Len(t, u1, 1)

	// Removing the last subscription prunes the user entirely.
	require.NoError(t, r.Delete(ctx, "u1", "https://e/2"))
	all, _ := r.All(ctx)
	assert.NotContains(t, all, "u1")

	// Deleting a missing subscription is a no-op.
	require.NoError(t, r.Delete(ctx, "nobody", "https://e/x"))
}
