// Package task holds the personal task entity for "My Day" (spec §5.7).
package task

import (
	"errors"
	"strings"
	"time"
)

// Priority, Status, and Source enumerations.
type (
	Priority string
	Status   string
	Source   string
)

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"

	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"

	SourceManual Source = "manual"
	SourceBrief  Source = "brief"
	SourceAlert  Source = "alert"
)

// Valid reports whether p is a known priority.
func (p Priority) Valid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	default:
		return false
	}
}

// Valid reports whether s is a known status.
func (s Status) Valid() bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

// Valid reports whether s is a known source.
func (s Source) Valid() bool {
	switch s {
	case SourceManual, SourceBrief, SourceAlert:
		return true
	default:
		return false
	}
}

// Task is a personal action item. FacilityID is optional (network-wide tasks).
type Task struct {
	ID         string
	Title      string
	Detail     string
	FacilityID string
	Priority   Priority
	Status     Status
	DueDate    time.Time
	AssignedTo string
	CreatedBy  string
	Source     Source
	CreatedAt  time.Time
}

// Validation errors.
var (
	ErrEmptyID         = errors.New("task: id is required")
	ErrEmptyTitle      = errors.New("task: title is required")
	ErrInvalidPriority = errors.New("task: invalid priority")
	ErrInvalidStatus   = errors.New("task: invalid status")
	ErrInvalidSource   = errors.New("task: invalid source")
)

// New validates and returns a Task.
func New(t Task) (Task, error) {
	t.ID = strings.TrimSpace(t.ID)
	t.Title = strings.TrimSpace(t.Title)
	switch {
	case t.ID == "":
		return Task{}, ErrEmptyID
	case t.Title == "":
		return Task{}, ErrEmptyTitle
	case !t.Priority.Valid():
		return Task{}, ErrInvalidPriority
	case !t.Status.Valid():
		return Task{}, ErrInvalidStatus
	case !t.Source.Valid():
		return Task{}, ErrInvalidSource
	}
	return t, nil
}
