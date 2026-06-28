package app_test

import (
	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/user"
)

// execPrincipal is a network-wide executive (sees everything).
func execPrincipal() auth.Principal {
	return auth.Principal{UserID: "u-exec", Name: "Sammy Adjei", Role: user.RoleExecutive}
}

// managerPrincipal is a facility manager scoped to facilityID.
func managerPrincipal(facilityID string) auth.Principal {
	return auth.Principal{UserID: "u-mgr", Name: "Ama Owusu", Role: user.RoleFacilityManager, FacilityID: facilityID}
}
