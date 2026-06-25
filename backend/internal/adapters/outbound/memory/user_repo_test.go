package memory_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/ports"
)

func TestUserRepoFindByEmail(t *testing.T) {
	u, err := user.New(user.User{ID: "u1", Name: "Sammy", Role: user.RoleExecutive})
	require.NoError(t, err)
	repo := memory.NewUserRepo(ports.Account{User: u, Email: "CEO@Gigmann.health", PasswordHash: "hash"})

	got, err := repo.FindByEmail(context.Background(), "ceo@gigmann.health") // case-insensitive
	require.NoError(t, err)
	assert.Equal(t, "u1", got.User.ID)

	_, err = repo.FindByEmail(context.Background(), "nobody@gigmann.health")
	assert.ErrorIs(t, err, ports.ErrAccountNotFound)
}
