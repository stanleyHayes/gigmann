package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/task"
	"github.com/xcreativs/gigmann/internal/ports"
)

// NewTaskInput describes a task to create (e.g. from a brief item or alert).
type NewTaskInput struct {
	Title      string
	Detail     string
	FacilityID string
	AssignedTo string
	Priority   task.Priority
	Source     task.Source
	DueDate    time.Time
}

// TaskService is the "My Day" use case: list tasks and update their status.
type TaskService struct {
	repo ports.TaskRepository
}

// NewTaskService wires the task use case to its repository.
func NewTaskService(repo ports.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

// List returns the tasks the principal may see: executives see all; a facility
// manager sees only their own facility's tasks (no IDOR).
func (s *TaskService) List(ctx context.Context, p auth.Principal) ([]task.Task, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("app: list tasks: %w", err)
	}
	if p.IsExecutive() {
		return items, nil
	}
	scoped := make([]task.Task, 0, len(items))
	for _, t := range items {
		if p.CanAccessFacility(t.FacilityID) {
			scoped = append(scoped, t)
		}
	}
	return scoped, nil
}

// UpdateStatus moves a task to a new status (todo/in_progress/done). A facility
// manager may only update their own facility's tasks (ErrForbidden otherwise).
func (s *TaskService) UpdateStatus(ctx context.Context, p auth.Principal, id string, status task.Status) (task.Task, error) {
	current, err := s.repo.Get(ctx, id)
	if err != nil {
		return task.Task{}, err
	}
	if !p.CanAccessFacility(current.FacilityID) {
		return task.Task{}, ErrForbidden
	}
	current.Status = status
	if err := s.repo.Save(ctx, current); err != nil {
		return task.Task{}, fmt.Errorf("app: save task: %w", err)
	}
	return current, nil
}

// Create makes a new "My Day" task (status todo). The source records where it came
// from (manual/brief/alert) for traceability. A facility manager may only create
// tasks scoped to their own facility (ErrForbidden otherwise).
func (s *TaskService) Create(ctx context.Context, p auth.Principal, in NewTaskInput) (task.Task, error) {
	if in.FacilityID != "" && !p.CanAccessFacility(in.FacilityID) {
		return task.Task{}, ErrForbidden
	}
	id, err := newTaskID()
	if err != nil {
		return task.Task{}, err
	}
	t, err := task.New(task.Task{
		ID: id, Title: in.Title, Detail: in.Detail, FacilityID: in.FacilityID,
		AssignedTo: in.AssignedTo, Priority: in.Priority, Status: task.StatusTodo,
		Source: in.Source, DueDate: in.DueDate, CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return task.Task{}, fmt.Errorf("app: new task: %w", err)
	}
	if err := s.repo.Save(ctx, t); err != nil {
		return task.Task{}, fmt.Errorf("app: save task: %w", err)
	}
	return t, nil
}

func newTaskID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("app: task id: %w", err)
	}
	return "task-" + hex.EncodeToString(b), nil
}
