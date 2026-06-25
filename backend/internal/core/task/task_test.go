package task_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/core/task"
)

func valid() task.Task {
	return task.Task{
		ID: "t1", Title: "Message Tafo manager about claims", Priority: task.PriorityHigh,
		Status: task.StatusTodo, Source: task.SourceBrief, FacilityID: "tafo-maternity",
	}
}

func TestNewValid(t *testing.T) {
	got, err := task.New(valid())
	require.NoError(t, err)
	assert.Equal(t, task.SourceBrief, got.Source)
}

func TestNewInvariants(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(tk *task.Task)
		wantErr error
	}{
		{"empty id", func(tk *task.Task) { tk.ID = "" }, task.ErrEmptyID},
		{"empty title", func(tk *task.Task) { tk.Title = "  " }, task.ErrEmptyTitle},
		{"bad priority", func(tk *task.Task) { tk.Priority = "urgent" }, task.ErrInvalidPriority},
		{"bad status", func(tk *task.Task) { tk.Status = "blocked" }, task.ErrInvalidStatus},
		{"bad source", func(tk *task.Task) { tk.Source = "email" }, task.ErrInvalidSource},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := valid()
			tt.mutate(&tk)
			_, err := task.New(tk)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestEnumValidity(t *testing.T) {
	assert.True(t, task.PriorityHigh.Valid())
	assert.False(t, task.Priority("x").Valid())
	assert.True(t, task.StatusDone.Valid())
	assert.False(t, task.Status("x").Valid())
	assert.True(t, task.SourceAlert.Valid())
	assert.False(t, task.Source("x").Valid())
}
