package alert_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/core/alert"
	"github.com/xcreativs/gigmann/internal/core/severity"
)

func valid() alert.Alert {
	return alert.Alert{
		ID: "a1", FacilityID: "tafo-maternity", Type: "revenue_drop",
		Severity: severity.Critical, Title: "Revenue down 22%", Detail: "Claims not submitted",
		Status: alert.StatusOpen,
	}
}

func TestNewValid(t *testing.T) {
	a, err := alert.New(valid())
	require.NoError(t, err)
	assert.Equal(t, severity.Critical, a.Severity)
}

func TestNewInvariants(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(a *alert.Alert)
		wantErr error
	}{
		{"empty id", func(a *alert.Alert) { a.ID = "" }, alert.ErrEmptyID},
		{"empty facility", func(a *alert.Alert) { a.FacilityID = "" }, alert.ErrEmptyFacilityID},
		{"empty type", func(a *alert.Alert) { a.Type = "" }, alert.ErrEmptyType},
		{"empty title", func(a *alert.Alert) { a.Title = " " }, alert.ErrEmptyTitle},
		{"bad severity", func(a *alert.Alert) { a.Severity = "meh" }, alert.ErrInvalidSeverity},
		{"bad status", func(a *alert.Alert) { a.Status = "weird" }, alert.ErrInvalidStatus},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := valid()
			tt.mutate(&a)
			_, err := alert.New(a)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestTransitions(t *testing.T) {
	a, err := alert.New(valid())
	require.NoError(t, err)

	dismissed, err := a.Dismiss()
	require.NoError(t, err)
	assert.Equal(t, alert.StatusDismissed, dismissed.Status)

	_, err = dismissed.Resolve()
	require.ErrorIs(t, err, alert.ErrAlreadyTerminal)

	resolved, err := a.Resolve()
	require.NoError(t, err)
	assert.Equal(t, alert.StatusResolved, resolved.Status)

	assert.True(t, alert.StatusOpen.Valid())
	assert.False(t, alert.Status("x").Valid())
}
