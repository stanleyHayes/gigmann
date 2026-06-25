package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xcreativs/gigmann/internal/adapters/inbound/httpapi"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/payer"
	"github.com/xcreativs/gigmann/internal/core/severity"
)

type errRepo struct{}

func (errRepo) List(context.Context) ([]facility.Facility, error) {
	return nil, errors.New("db down")
}

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	mix, _ := payer.New(70, 25, 5)
	f, err := facility.New(facility.Params{
		ID: "f1", Name: "Kasoa Polyclinic", Region: "Central", Town: "Kasoa",
		Type: "OPD", Beds: 40, Lifecycle: facility.LifecycleActive, Health: severity.Good,
		ManagerName: "Ama Owusu", PayerMix: mix,
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	repo := memory.NewFacilityRepo(f)
	return httpapi.NewRouter(app.NewFacilityService(repo))
}

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	newTestServer(t).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
}

func TestListFacilities(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/facilities", nil)
	newTestServer(t).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("want application/json, got %q", ct)
	}

	var body struct {
		Facilities []map[string]any `json:"facilities"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Facilities) != 1 {
		t.Fatalf("want 1 facility, got %d", len(body.Facilities))
	}
	if body.Facilities[0]["name"] != "Kasoa Polyclinic" {
		t.Errorf("unexpected name: %v", body.Facilities[0]["name"])
	}
}

func TestListFacilitiesError(t *testing.T) {
	h := httpapi.NewRouter(app.NewFacilityService(errRepo{}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/facilities", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", rec.Code)
	}
}
