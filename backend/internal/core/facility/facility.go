// Package facility holds the Facility domain entity and its value objects.
// It is pure domain code: no framework, database, or adapter imports.
package facility

import (
	"errors"
	"strings"
)

// Status is the AI-assessed operational health signal for a facility.
type Status string

const (
	StatusGood     Status = "good"
	StatusWatch    Status = "watch"
	StatusCritical Status = "critical"
)

// Valid reports whether s is a known status.
func (s Status) Valid() bool {
	switch s {
	case StatusGood, StatusWatch, StatusCritical:
		return true
	default:
		return false
	}
}

// Region is a Ghanaian administrative region (e.g. "Ashanti", "Central").
type Region string

// Facility is a single hospital, clinic, or diagnostic centre in the network.
type Facility struct {
	ID     string
	Name   string
	Region Region
	Town   string
	Beds   int
	Status Status
}

// Domain validation errors.
var (
	ErrEmptyID       = errors.New("facility: id is required")
	ErrEmptyName     = errors.New("facility: name is required")
	ErrEmptyRegion   = errors.New("facility: region is required")
	ErrNegativeBeds  = errors.New("facility: beds must be >= 0")
	ErrInvalidStatus = errors.New("facility: invalid status")
)

// New constructs a Facility, enforcing domain invariants.
func New(id, name string, region Region, town string, beds int, status Status) (Facility, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	switch {
	case id == "":
		return Facility{}, ErrEmptyID
	case name == "":
		return Facility{}, ErrEmptyName
	case strings.TrimSpace(string(region)) == "":
		return Facility{}, ErrEmptyRegion
	case beds < 0:
		return Facility{}, ErrNegativeBeds
	case !status.Valid():
		return Facility{}, ErrInvalidStatus
	}
	return Facility{
		ID:     id,
		Name:   name,
		Region: region,
		Town:   strings.TrimSpace(town),
		Beds:   beds,
		Status: status,
	}, nil
}
