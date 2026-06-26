package app

import (
	"context"
	"fmt"

	"github.com/xcreativs/gigmann/internal/core/task"
	"github.com/xcreativs/gigmann/internal/ports"
)

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
