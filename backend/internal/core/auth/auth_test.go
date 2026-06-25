package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/user"
)

func TestExecutiveAccessesAnyFacility(t *testing.T) {
	p := auth.Principal{UserID: "u1", Name: "Sammy", Role: user.RoleExecutive}
	assert.True(t, p.IsExecutive())
	assert.True(t, p.CanAccessFacility("kasoa"))
	assert.True(t, p.CanAccessFacility("tafo"))
}

func TestManagerScopedToOwnFacility(t *testing.T) {
	p := auth.Principal{UserID: "u2", Name: "Ama", Role: user.RoleFacilityManager, FacilityID: "kasoa"}
	assert.False(t, p.IsExecutive())
	assert.True(t, p.CanAccessFacility("kasoa"))
	assert.False(t, p.CanAccessFacility("tafo"))
}

func TestManagerWithoutFacilityAccessesNothing(t *testing.T) {
	p := auth.Principal{UserID: "u3", Role: user.RoleFacilityManager}
	assert.False(t, p.CanAccessFacility(""))
	assert.False(t, p.CanAccessFacility("kasoa"))
}
