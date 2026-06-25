package approval_test

import (
	"errors"
	"testing"
	"time"

	"github.com/xcreativs/gigmann/internal/core/approval"
	"github.com/xcreativs/gigmann/internal/core/money"
)

func valid() approval.Approval {
	return approval.Approval{
		ID: "ap1", Type: approval.TypeCapital, FacilityID: "assin-fosu",
		Amount: money.FromCedis(85000, 0), Title: "Ultrasound machine", RequestedBy: "Dr. Mensah",
		Status: approval.StatusPending,
	}
}

func TestNewValid(t *testing.T) {
	a, err := approval.New(valid())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if a.Amount.Pesewas() != 8500000 {
		t.Errorf("amount not set: %d", a.Amount.Pesewas())
	}
}

func TestNewInvariants(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(a *approval.Approval)
		wantErr error
	}{
		{"empty id", func(a *approval.Approval) { a.ID = "" }, approval.ErrEmptyID},
		{"empty title", func(a *approval.Approval) { a.Title = " " }, approval.ErrEmptyTitle},
		{"bad type", func(a *approval.Approval) { a.Type = "loan" }, approval.ErrInvalidType},
		{"bad status", func(a *approval.Approval) { a.Status = "maybe" }, approval.ErrInvalidStatus},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := valid()
			tt.mutate(&a)
			if _, err := approval.New(a); !errors.Is(err, tt.wantErr) {
				t.Fatalf("want %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestDecide(t *testing.T) {
	at := time.Date(2026, 6, 9, 8, 0, 0, 0, time.UTC)
	a, _ := approval.New(valid())

	approved, err := a.Decide(true, "Go ahead", at)
	if err != nil || approved.Status != approval.StatusApproved || approved.DecisionNote != "Go ahead" {
		t.Fatalf("approve failed: %v %+v", err, approved)
	}
	if _, err := approved.Decide(false, "", at); !errors.Is(err, approval.ErrAlreadyDecided) {
		t.Errorf("want ErrAlreadyDecided, got %v", err)
	}

	declined, err := a.Decide(false, "Defer to Q3", at)
	if err != nil || declined.Status != approval.StatusDeclined {
		t.Fatalf("decline failed: %v %+v", err, declined)
	}
}
