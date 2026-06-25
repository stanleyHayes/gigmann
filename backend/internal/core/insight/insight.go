// Package insight holds the Insight entity — AI-generated facility/network notes (spec §7).
package insight

import (
	"errors"
	"strings"
	"time"
)

// Insight is a generated note attached to a facility or the network.
type Insight struct {
	ID                string
	Type              string
	FacilityID        string
	Content           string
	SupportingFigures map[string]any
	GeneratedAt       time.Time
}

// Validation errors.
var (
	ErrEmptyID      = errors.New("insight: id is required")
	ErrEmptyType    = errors.New("insight: type is required")
	ErrEmptyContent = errors.New("insight: content is required")
)

// New validates and returns an Insight.
func New(i Insight) (Insight, error) {
	i.ID = strings.TrimSpace(i.ID)
	i.Type = strings.TrimSpace(i.Type)
	i.Content = strings.TrimSpace(i.Content)
	switch {
	case i.ID == "":
		return Insight{}, ErrEmptyID
	case i.Type == "":
		return Insight{}, ErrEmptyType
	case i.Content == "":
		return Insight{}, ErrEmptyContent
	}
	return i, nil
}
