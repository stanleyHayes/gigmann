package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xcreativs/gigmann/internal/adapters/inbound/httpapi"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/facility"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	f, err := facility.New("f1", "Kasoa Polyclinic", "Central", "Kasoa", 40, facility.StatusGood)
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
