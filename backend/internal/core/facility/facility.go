// Package facility holds the Facility aggregate and its value objects.
// Pure domain code: no framework, database, or adapter imports.
package facility

import (
	"errors"
	"strings"

	"github.com/xcreativs/gigmann/internal/core/payer"
	"github.com/xcreativs/gigmann/internal/core/severity"
)

// Lifecycle is a facility's operational lifecycle (spec §7 facilities.status).
type Lifecycle string

const (
	LifecycleActive   Lifecycle = "active"
	LifecycleRamping  Lifecycle = "ramping"
	LifecycleFlagship Lifecycle = "flagship"
)

// Valid reports whether l is a known lifecycle value.
func (l Lifecycle) Valid() bool {
	switch l {
	case LifecycleActive, LifecycleRamping, LifecycleFlagship:
		return true
	default:
		return false
	}
}

// Region is a Ghanaian administrative region (e.g. "Ashanti", "Central").
type Region string

// Facility is a hospital, clinic, or diagnostic centre in the network.
type Facility struct {
	ID          string
	Name        string
	Region      Region
	Town        string
	Type        string
	Beds        int
	Lifecycle   Lifecycle
	Health      severity.Severity // latest AI-assessed health (good/watch/critical)
	ManagerName string
	PayerMix    payer.Mix
	Latitude    float64
	Longitude   float64
}

// Params carries the inputs to New.
type Params struct {
	ID          string
	Name        string
	Region      Region
	Town        string
	Type        string
	Beds        int
	Lifecycle   Lifecycle
	Health      severity.Severity
	ManagerName string
	PayerMix    payer.Mix
	Latitude    float64
	Longitude   float64
}

// Domain validation errors.
var (
	ErrEmptyID          = errors.New("facility: id is required")
	ErrEmptyName        = errors.New("facility: name is required")
	ErrEmptyRegion      = errors.New("facility: region is required")
	ErrNegativeBeds     = errors.New("facility: beds must be >= 0")
	ErrInvalidLifecycle = errors.New("facility: invalid lifecycle")
	ErrInvalidHealth    = errors.New("facility: invalid health")
	ErrInvalidPayerMix  = errors.New("facility: invalid payer mix")
)

// New constructs a Facility, enforcing domain invariants.
func New(p Params) (Facility, error) {
	id := strings.TrimSpace(p.ID)
	name := strings.TrimSpace(p.Name)
	switch {
	case id == "":
		return Facility{}, ErrEmptyID
	case name == "":
		return Facility{}, ErrEmptyName
	case strings.TrimSpace(string(p.Region)) == "":
		return Facility{}, ErrEmptyRegion
	case p.Beds < 0:
		return Facility{}, ErrNegativeBeds
	case !p.Lifecycle.Valid():
		return Facility{}, ErrInvalidLifecycle
	case !p.Health.Valid():
		return Facility{}, ErrInvalidHealth
	case !p.PayerMix.Valid():
		return Facility{}, ErrInvalidPayerMix
	}
	return Facility{
		ID:          id,
		Name:        name,
		Region:      p.Region,
		Town:        strings.TrimSpace(p.Town),
		Type:        strings.TrimSpace(p.Type),
		Beds:        p.Beds,
		Lifecycle:   p.Lifecycle,
		Health:      p.Health,
		ManagerName: strings.TrimSpace(p.ManagerName),
		PayerMix:    p.PayerMix,
		Latitude:    p.Latitude,
		Longitude:   p.Longitude,
	}, nil
}
