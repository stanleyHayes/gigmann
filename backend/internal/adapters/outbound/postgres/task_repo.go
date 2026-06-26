package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/postgres/sqlcgen"
	"github.com/xcreativs/gigmann/internal/core/task"
	"github.com/xcreativs/gigmann/internal/ports"
)

// TaskRepo is a PostgreSQL implementation of ports.TaskRepository.
type TaskRepo struct {
	q *sqlcgen.Queries
}

var _ ports.TaskRepository = (*TaskRepo)(nil)

// NewTaskRepo builds a TaskRepo over a pgx pool (or any sqlcgen.DBTX).
func NewTaskRepo(db sqlcgen.DBTX) *TaskRepo {
	return &TaskRepo{q: sqlcgen.New(db)}
}

// List returns all tasks ordered by creation, mapped to the domain model.
func (r *TaskRepo) List(ctx context.Context) ([]task.Task, error) {
	rows, err := r.q.ListTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres: list tasks: %w", err)
	}
	out := make([]task.Task, 0, len(rows))
	for _, row := range rows {
		t, ferr := taskFromModel(row)
		if ferr != nil {
			return nil, fmt.Errorf("postgres: map task %q: %w", row.ID, ferr)
		}
		out = append(out, t)
	}
	return out, nil
}

// Get returns the task with the given id, or ErrTaskNotFound.
func (r *TaskRepo) Get(ctx context.Context, id string) (task.Task, error) {
	row, err := r.q.GetTask(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return task.Task{}, ports.ErrTaskNotFound
	}
	if err != nil {
		return task.Task{}, fmt.Errorf("postgres: get task: %w", err)
	}
	return taskFromModel(row)
}

// Save upserts a task.
func (r *TaskRepo) Save(ctx context.Context, t task.Task) error {
	if err := r.q.UpsertTask(ctx, taskParams(t)); err != nil {
		return fmt.Errorf("postgres: upsert task %q: %w", t.ID, err)
	}
	return nil
}

func taskParams(t task.Task) sqlcgen.UpsertTaskParams {
	return sqlcgen.UpsertTaskParams{
		ID:         t.ID,
		Title:      t.Title,
		Detail:     t.Detail,
		FacilityID: nullableStr(t.FacilityID),
		Priority:   string(t.Priority),
		Status:     string(t.Status),
		DueDate:    tsOptional(t.DueDate),
		AssignedTo: t.AssignedTo,
		CreatedBy:  t.CreatedBy,
		Source:     string(t.Source),
		CreatedAt:  tsRequired(t.CreatedAt),
	}
}

func taskFromModel(m sqlcgen.Task) (task.Task, error) {
	return task.New(task.Task{
		ID:         m.ID,
		Title:      m.Title,
		Detail:     m.Detail,
		FacilityID: derefStr(m.FacilityID),
		Priority:   task.Priority(m.Priority),
		Status:     task.Status(m.Status),
		DueDate:    timeFromTS(m.DueDate),
		AssignedTo: m.AssignedTo,
		CreatedBy:  m.CreatedBy,
		Source:     task.Source(m.Source),
		CreatedAt:  timeFromTS(m.CreatedAt),
	})
}
