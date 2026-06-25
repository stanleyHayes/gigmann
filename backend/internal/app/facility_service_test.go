package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/payer"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/ports/mocks"
)

func mustFacility(t *testing.T) facility.Facility {
	t.Helper()
	mix, err := payer.New(70, 25, 5)
	require.NoError(t, err)
	f, err := facility.New(facility.Params{
		ID: "f1", Name: "Kasoa", Region: "Central", Town: "Kasoa",
		Type: "OPD", Beds: 40, Lifecycle: facility.LifecycleActive, Health: severity.Good,
		ManagerName: "Ama Owusu", PayerMix: mix,
	})
	require.NoError(t, err)
	return f
}

func TestFacilityServiceList(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFacilityRepository(ctrl)
	repo.EXPECT().List(gomock.Any()).Return([]facility.Facility{mustFacility(t)}, nil)

	got, err := app.NewFacilityService(repo).List(context.Background())

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "f1", got[0].ID)
}

func TestFacilityServiceListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFacilityRepository(ctrl)
	repo.EXPECT().List(gomock.Any()).Return(nil, errors.New("boom"))

	_, err := app.NewFacilityService(repo).List(context.Background())

	require.Error(t, err)
}
