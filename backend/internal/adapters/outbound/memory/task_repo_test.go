package memory_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/core/task"
	"github.com/xcreativs/gigmann/internal/ports"
)

func todo(id string) task.Task {
	return task.Task{ID: id, Title: "T", Priority: task.PriorityMedium, Status: task.StatusTodo, Source: task.SourceManual}
}

func TestTaskRepoListGetSave(t *testing.T) {
	repo := memory.NewTaskRepo(todo("t1"), todo("t2"))

	all, err := repo.List(context.Background())
	require.NoError(t, err)
	require.Len(t, all, 2)

	got, err := repo.Get(context.Background(), "t2")
	require.NoError(t, err)
	got.Status = task.StatusDone
	require.NoError(t, repo.Save(context.Background(), got))

	again, err := repo.Get(context.Background(), "t2")
	require.NoError(t, err)
	assert.Equal(t, task.StatusDone, again.Status)
}

func TestTaskRepoGetNotFound(t *testing.T) {
	repo := memory.NewTaskRepo()
	_, err := repo.Get(context.Background(), "nope")
	assert.ErrorIs(t, err, ports.ErrTaskNotFound)
}
