package httpapi_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/xcreativs/gigmann/internal/adapters/inbound/httpapi"
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
		ID: "f1", Name: "Kasoa Polyclinic", Region: "Central", Town: "Kasoa",
		Type: "OPD", Beds: 40, Lifecycle: facility.LifecycleActive, Health: severity.Good,
		ManagerName: "Ama Owusu", PayerMix: mix,
	})
	require.NoError(t, err)
	return f
}

func serve(repo *mocks.MockFacilityRepository, method, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, nil)
	httpapi.NewRouter(app.NewFacilityService(repo)).ServeHTTP(rec, req)
	return rec
}

func TestHealthz(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFacilityRepository(ctrl)

	rec := serve(repo, http.MethodGet, "/healthz")

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestListFacilities(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFacilityRepository(ctrl)
	repo.EXPECT().List(gomock.Any()).Return([]facility.Facility{mustFacility(t)}, nil)

	rec := serve(repo, http.MethodGet, "/api/v1/facilities")

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body struct {
		Facilities []map[string]any `json:"facilities"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Len(t, body.Facilities, 1)
	assert.Equal(t, "Kasoa Polyclinic", body.Facilities[0]["name"])
}

func TestListFacilitiesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFacilityRepository(ctrl)
	repo.EXPECT().List(gomock.Any()).Return(nil, errors.New("db down"))

	rec := serve(repo, http.MethodGet, "/api/v1/facilities")

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
