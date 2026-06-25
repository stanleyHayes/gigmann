package staff_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	require.NoError(t, err)
	assert.Equal(t, "Physician Assistant", m.Role)
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
			_, err := staff.New(m)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestLicenceExpiringWithin(t *testing.T) {
	asOf := time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC)
	m := valid()
	assert.True(t, m.LicenceExpiringWithin(asOf, 30))
	assert.False(t, m.LicenceExpiringWithin(asOf, 5))
	m.LicenceExpiry = time.Time{}
	assert.False(t, m.LicenceExpiringWithin(asOf, 365))
}
