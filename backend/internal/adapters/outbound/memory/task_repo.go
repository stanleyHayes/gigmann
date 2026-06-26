package memory

import (
	"context"
	"sync"

	"github.com/xcreativs/gigmann/internal/core/task"
	"github.com/xcreativs/gigmann/internal/ports"
)

// TaskRepo is an in-memory ports.TaskRepository preserving seed order.
type TaskRepo struct {
	mu    sync.RWMutex
	items []task.Task
}

// NewTaskRepo creates a repository optionally seeded with tasks.
func NewTaskRepo(seed ...task.Task) *TaskRepo {
	return &TaskRepo{items: append([]task.Task{}, seed...)}
}

var _ ports.TaskRepository = (*TaskRepo)(nil)

// List returns a copy of all tasks.
func (r *TaskRepo) List(_ context.Context) ([]task.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]task.Task, len(r.items))
	copy(out, r.items)
	return out, nil
}

// Get returns the task with the given id, or ErrTaskNotFound.
func (r *TaskRepo) Get(_ context.Context, id string) (task.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, t := range r.items {
		if t.ID == id {
			return t, nil
		}
	}
	return task.Task{}, ports.ErrTaskNotFound
}

// Save updates an existing task in place (or appends a new one).
func (r *TaskRepo) Save(_ context.Context, t task.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == t.ID {
			r.items[i] = t
			return nil
		}
	}
	r.items = append(r.items, t)
	return nil
}
