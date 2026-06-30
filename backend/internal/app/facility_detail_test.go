package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/seed"
)

func TestFacilityDetail(t *testing.T) {
	net := seed.Generate(7, time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC), 14)
	require.NotEmpty(t, net.Facilities)
	svc := app.NewFacilityDetailService(net.Facilities, net.Inventory, net.Staff, net.Alerts, net.Metrics)

	id := net.Facilities[0].ID
	d, err := svc.Detail(context.Background(), execPrincipal(), id)
	require.NoError(t, err)
	assert.Equal(t, id, d.Facility.ID)
	assert.NotEmpty(t, d.KPIs)
	// every returned sub-resource belongs to the requested facility
	for _, it := range d.Inventory {
		assert.Equal(t, id, it.FacilityID)
	}
	for _, m := range d.Staff {
		assert.Equal(t, id, m.FacilityID)
	}
	for _, a := range d.Alerts {
		assert.Equal(t, id, a.FacilityID)
	}
}

func TestFacilityDetailNotFound(t *testing.T) {
	svc := app.NewFacilityDetailService(nil, nil, nil, nil)
	_, err := svc.Detail(context.Background(), execPrincipal(), "ghost")
	assert.ErrorIs(t, err, app.ErrFacilityNotFound)
}

func TestFacilityDetailScopesManager(t *testing.T) {
	net := seed.Generate(7, time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC), 14)
	svc := app.NewFacilityDetailService(net.Facilities, net.Inventory, net.Staff, net.Alerts)
	own := net.Facilities[0].ID
	other := net.Facilities[1].ID

	_, err := svc.Detail(context.Background(), managerPrincipal(own), own)
	require.NoError(t, err, "a manager may drill into their own facility")

	_, err = svc.Detail(context.Background(), managerPrincipal(own), other)
	require.ErrorIs(t, err, app.ErrForbidden, "a manager may not drill into another facility (IDOR)")
}
