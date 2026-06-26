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

func TestUserRepoFindByIDAndSave(t *testing.T) {
	u, err := user.New(user.User{ID: "u1", Name: "Sammy", Role: user.RoleExecutive})
	require.NoError(t, err)
	repo := memory.NewUserRepo(ports.Account{User: u, Email: "ceo@gigmann.health", PasswordHash: "h"})

	got, err := repo.FindByID(context.Background(), "u1")
	require.NoError(t, err)
	assert.Equal(t, "ceo@gigmann.health", got.Email)

	got.MFASecret = "SECRET"
	require.NoError(t, repo.Save(context.Background(), got))
	again, err := repo.FindByID(context.Background(), "u1")
	require.NoError(t, err)
	assert.Equal(t, "SECRET", again.MFASecret)

	_, err = repo.FindByID(context.Background(), "nobody")
	assert.ErrorIs(t, err, ports.ErrAccountNotFound)
}
