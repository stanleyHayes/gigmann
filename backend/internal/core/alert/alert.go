// Package alert holds the Alert entity for the attention feed (spec §5.5).
package alert

import (
	"errors"
	"strings"
	"time"

	"github.com/xcreativs/gigmann/internal/core/severity"
)

// Status is an alert's lifecycle state.
type Status string

const (
	StatusOpen      Status = "open"
	StatusDismissed Status = "dismissed"
	StatusResolved  Status = "resolved"
)

// Valid reports whether s is a known status.
func (s Status) Valid() bool {
	switch s {
	case StatusOpen, StatusDismissed, StatusResolved:
		return true
	default:
		return false
	}
}

// Alert is an exception/risk surfaced by the signal engine for a facility.
type Alert struct {
	ID                string
	FacilityID        string
	Type              string
	Severity          severity.Severity
	Title             string
	Detail            string
	SupportingFigures map[string]any
	Status            Status
	CreatedAt         time.Time
}

// Validation errors.
var (
	ErrEmptyID          = errors.New("alert: id is required")
	ErrEmptyFacilityID  = errors.New("alert: facility_id is required")
	ErrEmptyType        = errors.New("alert: type is required")
	ErrEmptyTitle       = errors.New("alert: title is required")
	ErrInvalidSeverity  = errors.New("alert: invalid severity")
	ErrInvalidStatus    = errors.New("alert: invalid status")
	ErrAlreadyTerminal  = errors.New("alert: already dismissed or resolved")
)

// New validates and returns an Alert.
func New(a Alert) (Alert, error) {
	a.ID = strings.TrimSpace(a.ID)
	a.FacilityID = strings.TrimSpace(a.FacilityID)
	a.Type = strings.TrimSpace(a.Type)
	a.Title = strings.TrimSpace(a.Title)
	switch {
	case a.ID == "":
		return Alert{}, ErrEmptyID
	case a.FacilityID == "":
		return Alert{}, ErrEmptyFacilityID
	case a.Type == "":
		return Alert{}, ErrEmptyType
	case a.Title == "":
		return Alert{}, ErrEmptyTitle
	case !a.Severity.Valid():
		return Alert{}, ErrInvalidSeverity
	case !a.Status.Valid():
		return Alert{}, ErrInvalidStatus
	}
	return a, nil
}

// Dismiss marks an open alert as dismissed.
func (a Alert) Dismiss() (Alert, error) { return a.transition(StatusDismissed) }

// Resolve marks an open alert as resolved.
func (a Alert) Resolve() (Alert, error) { return a.transition(StatusResolved) }

func (a Alert) transition(to Status) (Alert, error) {
	if a.Status != StatusOpen {
		return Alert{}, ErrAlreadyTerminal
	}
	a.Status = to
	return a, nil
}
