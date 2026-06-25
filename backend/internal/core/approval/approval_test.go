package approval_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	require.NoError(t, err)
	assert.Equal(t, int64(8500000), a.Amount.Pesewas())
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
			_, err := approval.New(a)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestDecide(t *testing.T) {
	at := time.Date(2026, 6, 9, 8, 0, 0, 0, time.UTC)
	a, err := approval.New(valid())
	require.NoError(t, err)

	approved, err := a.Decide(true, "Go ahead", at)
	require.NoError(t, err)
	assert.Equal(t, approval.StatusApproved, approved.Status)
	assert.Equal(t, "Go ahead", approved.DecisionNote)

	_, err = approved.Decide(false, "", at)
	require.ErrorIs(t, err, approval.ErrAlreadyDecided)

	declined, err := a.Decide(false, "Defer to Q3", at)
	require.NoError(t, err)
	assert.Equal(t, approval.StatusDeclined, declined.Status)
}
