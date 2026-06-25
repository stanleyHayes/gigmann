// Package user holds the User entity and roles (spec §7): executive vs facility_manager.
package user

import (
	"errors"
	"strings"
)

// Role is a user's access role.
type Role string

const (
	RoleExecutive       Role = "executive"
	RoleFacilityManager Role = "facility_manager"
)

// Valid reports whether r is a known role.
func (r Role) Valid() bool {
	switch r {
	case RoleExecutive, RoleFacilityManager:
		return true
	default:
		return false
	}
}

// Preferences captures personalisation (spec §5.12).
type Preferences struct {
	WatchedMetrics []string
	Thresholds     map[string]float64
}

// User is an authenticated cockpit user.
type User struct {
	ID          string
	Name        string
	Role        Role
	FacilityID  string // required for facility_manager; must be empty for executive
	Preferences Preferences
}

// Validation errors.
var (
	ErrEmptyID              = errors.New("user: id is required")
	ErrEmptyName            = errors.New("user: name is required")
	ErrInvalidRole          = errors.New("user: invalid role")
	ErrManagerNeedsFacility = errors.New("user: facility_manager must have a facility_id")
	ErrExecutiveHasFacility = errors.New("user: executive must not have a facility_id")
)

// New validates and returns a User, enforcing the role/facility relationship.
func New(u User) (User, error) {
	u.ID = strings.TrimSpace(u.ID)
	u.Name = strings.TrimSpace(u.Name)
	u.FacilityID = strings.TrimSpace(u.FacilityID)
	switch {
	case u.ID == "":
		return User{}, ErrEmptyID
	case u.Name == "":
		return User{}, ErrEmptyName
	case !u.Role.Valid():
		return User{}, ErrInvalidRole
	case u.Role == RoleFacilityManager && u.FacilityID == "":
		return User{}, ErrManagerNeedsFacility
	case u.Role == RoleExecutive && u.FacilityID != "":
		return User{}, ErrExecutiveHasFacility
	}
	return u, nil
}
