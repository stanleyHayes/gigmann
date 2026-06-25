package task_test

import (
	"errors"
	"testing"

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
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Source != task.SourceBrief {
		t.Errorf("source not set: %q", got.Source)
	}
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
			if _, err := task.New(tk); !errors.Is(err, tt.wantErr) {
				t.Fatalf("want %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestEnumValidity(t *testing.T) {
	if !task.PriorityHigh.Valid() || task.Priority("x").Valid() {
		t.Error("priority validity wrong")
	}
	if !task.StatusDone.Valid() || task.Status("x").Valid() {
		t.Error("status validity wrong")
	}
	if !task.SourceAlert.Valid() || task.Source("x").Valid() {
		t.Error("source validity wrong")
	}
}
