package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/core/user"
)

func TestNewExecutiveValid(t *testing.T) {
	u, err := user.New(user.User{ID: "u1", Name: "Sammy Adjei", Role: user.RoleExecutive})
	require.NoError(t, err)
	assert.Equal(t, user.RoleExecutive, u.Role)
}

func TestNewManagerValid(t *testing.T) {
	u, err := user.New(user.User{ID: "u2", Name: "Ama Owusu", Role: user.RoleFacilityManager, FacilityID: "kasoa"})
	require.NoError(t, err)
	assert.Equal(t, "kasoa", u.FacilityID)
}

func TestNewInvariants(t *testing.T) {
	tests := []struct {
		name    string
		in      user.User
		wantErr error
	}{
		{"empty id", user.User{Name: "X", Role: user.RoleExecutive}, user.ErrEmptyID},
		{"empty name", user.User{ID: "u", Role: user.RoleExecutive}, user.ErrEmptyName},
		{"bad role", user.User{ID: "u", Name: "X", Role: "admin"}, user.ErrInvalidRole},
		{"manager no facility", user.User{ID: "u", Name: "X", Role: user.RoleFacilityManager}, user.ErrManagerNeedsFacility},
		{"executive with facility", user.User{ID: "u", Name: "X", Role: user.RoleExecutive, FacilityID: "kasoa"}, user.ErrExecutiveHasFacility},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := user.New(tt.in)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestRoleValid(t *testing.T) {
	assert.True(t, user.RoleExecutive.Valid())
	assert.True(t, user.RoleFacilityManager.Valid())
	assert.False(t, user.Role("x").Valid())
}
