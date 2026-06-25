package staff_test

import (
	"errors"
	"testing"
	"time"

	"github.com/xcreativs/gigmann/internal/core/staff"
)

func valid() staff.Member {
	return staff.Member{
		ID: "pa-tamale-1", FacilityID: "tamale-north", Name: "Yaw Boateng", Role: "Physician Assistant",
		LicenceNumber: "PA-2024-117", LicenceExpiry: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		Status: "active", AttritionRisk: 0.6, JoinedDate: time.Date(2022, 1, 10, 0, 0, 0, 0, time.UTC),
	}
}

func TestNewValid(t *testing.T) {
	m, err := staff.New(valid())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if m.Role != "Physician Assistant" {
		t.Errorf("role not set: %q", m.Role)
	}
}

func TestNewInvariants(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(m *staff.Member)
		wantErr error
	}{
		{"empty id", func(m *staff.Member) { m.ID = "" }, staff.ErrEmptyID},
		{"empty facility", func(m *staff.Member) { m.FacilityID = "" }, staff.ErrEmptyFacilityID},
		{"empty name", func(m *staff.Member) { m.Name = "" }, staff.ErrEmptyName},
		{"empty role", func(m *staff.Member) { m.Role = " " }, staff.ErrEmptyRole},
		{"bad risk", func(m *staff.Member) { m.AttritionRisk = 1.4 }, staff.ErrBadRisk},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := valid()
			tt.mutate(&m)
			if _, err := staff.New(m); !errors.Is(err, tt.wantErr) {
				t.Fatalf("want %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestLicenceExpiringWithin(t *testing.T) {
	asOf := time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC)
	m := valid() // expiry 2026-07-01
	if !m.LicenceExpiringWithin(asOf, 30) {
		t.Error("licence expiring on 2026-07-01 should be within 30 days of 2026-06-09")
	}
	if m.LicenceExpiringWithin(asOf, 5) {
		t.Error("should not be within 5 days")
	}
	m.LicenceExpiry = time.Time{}
	if m.LicenceExpiringWithin(asOf, 365) {
		t.Error("zero expiry should never be 'expiring'")
	}
}
