package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

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

// List returns the current tasks.
func (s *TaskService) List(ctx context.Context) ([]task.Task, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("app: list tasks: %w", err)
	}
	return items, nil
}

// UpdateStatus moves a task to a new status (todo/in_progress/done).
func (s *TaskService) UpdateStatus(ctx context.Context, id string, status task.Status) (task.Task, error) {
	current, err := s.repo.Get(ctx, id)
	if err != nil {
		return task.Task{}, err
	}
	current.Status = status
	if err := s.repo.Save(ctx, current); err != nil {
		return task.Task{}, fmt.Errorf("app: save task: %w", err)
	}
	return current, nil
}

// Create makes a new "My Day" task (status todo). The source records where it came
// from (manual/brief/alert) for traceability.
func (s *TaskService) Create(ctx context.Context, in NewTaskInput) (task.Task, error) {
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
