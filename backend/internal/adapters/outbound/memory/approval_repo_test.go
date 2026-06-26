package memory_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/core/approval"
	"github.com/xcreativs/gigmann/internal/ports"
)

func pendingApproval(id string) approval.Approval {
	return approval.Approval{ID: id, Type: approval.TypeCapital, Title: "X", Status: approval.StatusPending}
}

func TestApprovalRepoListGetSave(t *testing.T) {
	repo := memory.NewApprovalRepo(pendingApproval("a1"), pendingApproval("a2"))

	all, err := repo.List(context.Background())
	require.NoError(t, err)
	require.Len(t, all, 2)

	got, err := repo.Get(context.Background(), "a2")
	require.NoError(t, err)
	assert.Equal(t, "a2", got.ID)

	got.Status = approval.StatusApproved
	require.NoError(t, repo.Save(context.Background(), got))
	again, err := repo.Get(context.Background(), "a2")
	require.NoError(t, err)
	assert.Equal(t, approval.StatusApproved, again.Status)
}

func TestApprovalRepoGetNotFound(t *testing.T) {
	repo := memory.NewApprovalRepo()
	_, err := repo.Get(context.Background(), "nope")
	assert.ErrorIs(t, err, ports.ErrApprovalNotFound)
}
