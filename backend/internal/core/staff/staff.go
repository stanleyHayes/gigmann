// Package staff holds the Staff member entity used by staff-signal detection (spec §6.3).
package staff

import (
	"errors"
	"strings"
	"time"
)

// Member is a staff member at a facility.
type Member struct {
	ID            string
	FacilityID    string
	Name          string
	Role          string
	LicenceNumber string
	LicenceExpiry time.Time
	Status        string
	AttritionRisk float64 // 0..1
	JoinedDate    time.Time
}

// Validation errors.
var (
	ErrEmptyID         = errors.New("staff: id is required")
	ErrEmptyFacilityID = errors.New("staff: facility_id is required")
	ErrEmptyName       = errors.New("staff: name is required")
	ErrEmptyRole       = errors.New("staff: role is required")
	ErrBadRisk         = errors.New("staff: attrition_risk must be within 0..1")
)

// New validates and returns a Member.
func New(m Member) (Member, error) {
	m.ID = strings.TrimSpace(m.ID)
	m.FacilityID = strings.TrimSpace(m.FacilityID)
	m.Name = strings.TrimSpace(m.Name)
	m.Role = strings.TrimSpace(m.Role)
	switch {
	case m.ID == "":
		return Member{}, ErrEmptyID
	case m.FacilityID == "":
		return Member{}, ErrEmptyFacilityID
	case m.Name == "":
		return Member{}, ErrEmptyName
	case m.Role == "":
		return Member{}, ErrEmptyRole
	case m.AttritionRisk < 0 || m.AttritionRisk > 1:
		return Member{}, ErrBadRisk
	}
	return m, nil
}

// LicenceExpiringWithin reports whether the licence expires within days of asOf.
func (m Member) LicenceExpiringWithin(asOf time.Time, days int) bool {
	if m.LicenceExpiry.IsZero() {
		return false
	}
	return m.LicenceExpiry.Before(asOf.AddDate(0, 0, days))
}
