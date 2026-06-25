package alert_test

import (
	"errors"
	"testing"

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
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if a.Severity != severity.Critical {
		t.Errorf("severity not set: %q", a.Severity)
	}
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
			if _, err := alert.New(a); !errors.Is(err, tt.wantErr) {
				t.Fatalf("want %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestTransitions(t *testing.T) {
	a, _ := alert.New(valid())
	dismissed, err := a.Dismiss()
	if err != nil || dismissed.Status != alert.StatusDismissed {
		t.Fatalf("dismiss failed: %v status=%q", err, dismissed.Status)
	}
	if _, err := dismissed.Resolve(); !errors.Is(err, alert.ErrAlreadyTerminal) {
		t.Errorf("want ErrAlreadyTerminal, got %v", err)
	}

	resolved, err := a.Resolve()
	if err != nil || resolved.Status != alert.StatusResolved {
		t.Fatalf("resolve failed: %v status=%q", err, resolved.Status)
	}

	if !alert.StatusOpen.Valid() || alert.Status("x").Valid() {
		t.Error("Status.Valid wrong")
	}
}
