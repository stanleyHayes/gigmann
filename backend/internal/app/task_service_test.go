package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/task"
	"github.com/xcreativs/gigmann/internal/ports"
	"github.com/xcreativs/gigmann/internal/ports/mocks"
)

func todoTask() task.Task {
	return task.Task{ID: "t1", Title: "Call Kasoa", Priority: task.PriorityHigh, Status: task.StatusTodo, Source: task.SourceBrief}
}

func TestTaskList(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTaskRepository(ctrl)
	repo.EXPECT().List(gomock.Any()).Return([]task.Task{todoTask()}, nil)

	got, err := app.NewTaskService(repo).List(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
}

func TestTaskUpdateStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTaskRepository(ctrl)
	repo.EXPECT().Get(gomock.Any(), "t1").Return(todoTask(), nil)
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

	out, err := app.NewTaskService(repo).UpdateStatus(context.Background(), "t1", task.StatusDone)
	require.NoError(t, err)
	assert.Equal(t, task.StatusDone, out.Status)
}

func TestTaskUpdateStatusNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTaskRepository(ctrl)
	repo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(task.Task{}, ports.ErrTaskNotFound)

	_, err := app.NewTaskService(repo).UpdateStatus(context.Background(), "missing", task.StatusDone)
	assert.ErrorIs(t, err, ports.ErrTaskNotFound)
}

func TestTaskServiceCreate(t *testing.T) {
	repo := memory.NewTaskRepo()
	svc := app.NewTaskService(repo)
	created, err := svc.Create(context.Background(), app.NewTaskInput{
		Title: "Message Tafo manager", FacilityID: "tafo-maternity",
		Priority: task.PriorityHigh, Source: task.SourceBrief,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, task.StatusTodo, created.Status)
	assert.Equal(t, task.SourceBrief, created.Source)

	list, err := svc.List(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 1)
}

func TestTaskServiceCreateEmptyTitle(t *testing.T) {
	svc := app.NewTaskService(memory.NewTaskRepo())
	_, err := svc.Create(context.Background(), app.NewTaskInput{Title: "  ", Priority: task.PriorityMedium, Source: task.SourceManual})
	require.ErrorIs(t, err, task.ErrEmptyTitle)
}
