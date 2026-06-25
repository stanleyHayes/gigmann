// Package brief holds the Daily Brief aggregate — the hero output (spec §5.1, §6.2).
package brief

import (
	"errors"
	"strings"
	"time"

	"github.com/xcreativs/gigmann/internal/core/severity"
)

// Item is one prioritised entry in the brief.
type Item struct {
	Severity         severity.Severity
	FacilityID       string
	Headline         string
	Explanation      string
	SuggestedActions []string
}

// Brief is the AI-narrated morning brief over computed signals.
type Brief struct {
	ID              string
	Date            time.Time
	Prose           string
	Items           []Item
	GeneratedAt     time.Time
	Model           string
	SourceSignalIDs []string
}

// Validation errors.
var (
	ErrEmptyID             = errors.New("brief: id is required")
	ErrZeroDate            = errors.New("brief: date is required")
	ErrInvalidItemSeverity = errors.New("brief: item has invalid severity")
	ErrEmptyHeadline       = errors.New("brief: item headline is required")
)

// New validates and returns a Brief, checking each item.
func New(b Brief) (Brief, error) {
	b.ID = strings.TrimSpace(b.ID)
	if b.ID == "" {
		return Brief{}, ErrEmptyID
	}
	if b.Date.IsZero() {
		return Brief{}, ErrZeroDate
	}
	for _, it := range b.Items {
		if !it.Severity.Valid() {
			return Brief{}, ErrInvalidItemSeverity
		}
		if strings.TrimSpace(it.Headline) == "" {
			return Brief{}, ErrEmptyHeadline
		}
	}
	return b, nil
}
