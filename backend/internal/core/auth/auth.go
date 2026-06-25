// Package auth holds the authenticated principal and pure authorization rules
// (spec §4.1, §7). Pure domain code: no I/O, no framework imports.
package auth

import "github.com/xcreativs/gigmann/internal/core/user"

// Principal is an authenticated identity carried through the request context.
type Principal struct {
	UserID     string
	Name       string
	Role       user.Role
	FacilityID string
}

// IsExecutive reports whether the principal has network-wide (executive) access.
func (p Principal) IsExecutive() bool { return p.Role == user.RoleExecutive }

// CanAccessFacility enforces facility scoping: executives see the whole network;
// facility managers may only access their own facility (guards against IDOR).
func (p Principal) CanAccessFacility(facilityID string) bool {
	if p.IsExecutive() {
		return true
	}
	return p.FacilityID != "" && p.FacilityID == facilityID
}
