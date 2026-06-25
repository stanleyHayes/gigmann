package httpapi_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/xcreativs/gigmann/internal/adapters/inbound/httpapi"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/payer"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/ports/mocks"
	"github.com/xcreativs/gigmann/internal/seed"
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

func serve(t *testing.T, repo *mocks.MockFacilityRepository, briefs *mocks.MockBriefGenerator, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, nil)
	metricsSvc := app.NewMetricsService(seed.Generate(7, time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC), 14).Metrics)
	httpapi.NewRouter(app.NewFacilityService(repo), metricsSvc, briefs).ServeHTTP(rec, req)
	return rec
}

func TestHealthz(t *testing.T) {
	ctrl := gomock.NewController(t)
	rec := serve(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl), http.MethodGet, "/healthz")
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestListFacilities(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFacilityRepository(ctrl)
	repo.EXPECT().List(gomock.Any()).Return([]facility.Facility{mustFacility(t)}, nil)

	rec := serve(t, repo, mocks.NewMockBriefGenerator(ctrl), http.MethodGet, "/api/v1/facilities")

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

	rec := serve(t, repo, mocks.NewMockBriefGenerator(ctrl), http.MethodGet, "/api/v1/facilities")
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetBrief(t *testing.T) {
	ctrl := gomock.NewController(t)
	briefs := mocks.NewMockBriefGenerator(ctrl)
	b, err := brief.New(brief.Brief{
		ID: "b-2026-06-09", Date: time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC), Prose: "Good morning, Sammy.",
		Items: []brief.Item{{
			Severity: severity.Critical, FacilityID: "tafo-maternity", Headline: "Tafo needs you first",
			Explanation: "claims not submitted", SuggestedActions: []string{"Why?", "Message the manager"},
		}},
		Model: "local-deterministic",
	})
	require.NoError(t, err)
	briefs.EXPECT().Generate(gomock.Any()).Return(b, nil)

	rec := serve(t, mocks.NewMockFacilityRepository(ctrl), briefs, http.MethodGet, "/api/v1/brief")

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		ID    string           `json:"id"`
		Prose string           `json:"prose"`
		Items []map[string]any `json:"items"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, "b-2026-06-09", body.ID)
	require.Len(t, body.Items, 1)
	assert.Equal(t, "critical", body.Items[0]["severity"])
	assert.Equal(t, "Tafo needs you first", body.Items[0]["headline"])
}

func TestGetBriefError(t *testing.T) {
	ctrl := gomock.NewController(t)
	briefs := mocks.NewMockBriefGenerator(ctrl)
	briefs.EXPECT().Generate(gomock.Any()).Return(brief.Brief{}, errors.New("api down"))

	rec := serve(t, mocks.NewMockFacilityRepository(ctrl), briefs, http.MethodGet, "/api/v1/brief")
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	rec := serve(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl), http.MethodGet, "/api/v1/metrics")

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		AsOf string           `json:"as_of"`
		KPIs []map[string]any `json:"kpis"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Len(t, body.KPIs, 4)
	assert.Equal(t, "revenue", body.KPIs[0]["key"])
	assert.NotEmpty(t, body.KPIs[0]["series"])
}
