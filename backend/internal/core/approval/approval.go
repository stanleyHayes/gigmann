// Package approval holds the Approval entity for decision routing (spec §5.8).
package approval

import (
	"errors"
	"strings"
	"time"

	"github.com/xcreativs/gigmann/internal/core/money"
)

// Type and Status enumerations.
type (
	Type   string
	Status string
)

const (
	TypeCapital Type = "capital"
	TypeHire    Type = "hire"
	TypeReorder Type = "reorder"

	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusDeclined Status = "declined"
)

// Valid reports whether t is a known approval type.
func (t Type) Valid() bool {
	switch t {
	case TypeCapital, TypeHire, TypeReorder:
		return true
	default:
		return false
	}
}

// Valid reports whether s is a known status.
func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusApproved, StatusDeclined:
		return true
	default:
		return false
	}
}

// Approval is a decision routed to the executive for sign-off.
type Approval struct {
	ID           string
	Type         Type
	FacilityID   string
	Amount       money.Cedis
	Title        string
	Context      string
	RequestedBy  string
	Status       Status
	DecidedAt    time.Time
	DecisionNote string
	CreatedAt    time.Time
}

// Validation errors.
var (
	ErrEmptyID        = errors.New("approval: id is required")
	ErrEmptyTitle     = errors.New("approval: title is required")
	ErrInvalidType    = errors.New("approval: invalid type")
	ErrInvalidStatus  = errors.New("approval: invalid status")
	ErrAlreadyDecided = errors.New("approval: already decided")
)

// New validates and returns an Approval.
func New(a Approval) (Approval, error) {
	a.ID = strings.TrimSpace(a.ID)
	a.Title = strings.TrimSpace(a.Title)
	switch {
	case a.ID == "":
		return Approval{}, ErrEmptyID
	case a.Title == "":
		return Approval{}, ErrEmptyTitle
	case !a.Type.Valid():
		return Approval{}, ErrInvalidType
	case !a.Status.Valid():
		return Approval{}, ErrInvalidStatus
	}
	return a, nil
}

// Decide records an approve/decline decision on a pending approval.
func (a Approval) Decide(approved bool, note string, at time.Time) (Approval, error) {
	if a.Status != StatusPending {
		return Approval{}, ErrAlreadyDecided
	}
	if approved {
		a.Status = StatusApproved
	} else {
		a.Status = StatusDeclined
	}
	a.DecisionNote = strings.TrimSpace(note)
	a.DecidedAt = at
	return a, nil
}
