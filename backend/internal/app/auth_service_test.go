package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/ports"
	"github.com/xcreativs/gigmann/internal/ports/mocks"
)

func execAccount(t *testing.T) ports.Account {
	t.Helper()
	u, err := user.New(user.User{ID: "u1", Name: "Sammy Adjei", Role: user.RoleExecutive})
	require.NoError(t, err)
	return ports.Account{User: u, Email: "ceo@gigmann.health", PasswordHash: "stored-hash"}
}

func TestLoginSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := mocks.NewMockUserRepository(ctrl)
	hasher := mocks.NewMockPasswordHasher(ctrl)
	tokens := mocks.NewMockTokenService(ctrl)

	users.EXPECT().FindByEmail(gomock.Any(), "ceo@gigmann.health").Return(execAccount(t), nil)
	hasher.EXPECT().Verify("pw", "stored-hash").Return(true, nil)
	tokens.EXPECT().Issue(gomock.Any()).Return("signed-token", nil)

	svc := app.NewAuthService(users, hasher, tokens)
	tok, p, err := svc.Login(context.Background(), "ceo@gigmann.health", "pw")

	require.NoError(t, err)
	assert.Equal(t, "signed-token", tok)
	assert.Equal(t, "u1", p.UserID)
	assert.Equal(t, user.RoleExecutive, p.Role)
}

func TestLoginUnknownEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := mocks.NewMockUserRepository(ctrl)
	users.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(ports.Account{}, ports.ErrAccountNotFound)

	svc := app.NewAuthService(users, mocks.NewMockPasswordHasher(ctrl), mocks.NewMockTokenService(ctrl))
	_, _, err := svc.Login(context.Background(), "nobody@gigmann.health", "pw")
	assert.ErrorIs(t, err, app.ErrInvalidCredentials)
}

func TestLoginWrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := mocks.NewMockUserRepository(ctrl)
	hasher := mocks.NewMockPasswordHasher(ctrl)

	users.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(execAccount(t), nil)
	hasher.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(false, nil)

	svc := app.NewAuthService(users, hasher, mocks.NewMockTokenService(ctrl))
	_, _, err := svc.Login(context.Background(), "ceo@gigmann.health", "bad")
	assert.ErrorIs(t, err, app.ErrInvalidCredentials)
}

func TestLoginTokenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	users := mocks.NewMockUserRepository(ctrl)
	hasher := mocks.NewMockPasswordHasher(ctrl)
	tokens := mocks.NewMockTokenService(ctrl)

	users.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(execAccount(t), nil)
	hasher.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(true, nil)
	tokens.EXPECT().Issue(gomock.Any()).Return("", errors.New("kms down"))

	svc := app.NewAuthService(users, hasher, tokens)
	_, _, err := svc.Login(context.Background(), "ceo@gigmann.health", "pw")
	require.Error(t, err)
	assert.NotErrorIs(t, err, app.ErrInvalidCredentials)
}
