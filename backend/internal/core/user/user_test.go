package user_test

import (
	"errors"
	"testing"

	"github.com/xcreativs/gigmann/internal/core/user"
)

func TestNewExecutiveValid(t *testing.T) {
	u, err := user.New(user.User{ID: "u1", Name: "Sammy Adjei", Role: user.RoleExecutive})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if u.Role != user.RoleExecutive {
		t.Errorf("role not set: %q", u.Role)
	}
}

func TestNewManagerValid(t *testing.T) {
	u, err := user.New(user.User{ID: "u2", Name: "Ama Owusu", Role: user.RoleFacilityManager, FacilityID: "kasoa"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if u.FacilityID != "kasoa" {
		t.Errorf("facility not set: %q", u.FacilityID)
	}
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
			if _, err := user.New(tt.in); !errors.Is(err, tt.wantErr) {
				t.Fatalf("want %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestRoleValid(t *testing.T) {
	if !user.RoleExecutive.Valid() || !user.RoleFacilityManager.Valid() || user.Role("x").Valid() {
		t.Error("role validity wrong")
	}
}
