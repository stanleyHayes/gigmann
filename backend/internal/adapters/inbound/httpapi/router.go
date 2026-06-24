// Package httpapi is the inbound HTTP adapter (Chi). Handlers are thin: they
// translate HTTP to/from application use cases and never hold business logic.
package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/facility"
)

// NewRouter builds the HTTP handler, wiring routes to application services.
func NewRouter(facilities *app.FacilityService) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))

	r.Get("/healthz", handleHealth)
	r.Get("/readyz", handleHealth)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/facilities", handleListFacilities(facilities))
	})

	return r
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type facilityDTO struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Region string `json:"region"`
	Town   string `json:"town"`
	Beds   int    `json:"beds"`
	Status string `json:"status"`
}

func handleListFacilities(svc *app.FacilityService) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		items, err := svc.List(req.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
			return
		}
		dtos := make([]facilityDTO, 0, len(items))
		for _, f := range items {
			dtos = append(dtos, toDTO(f))
		}
		writeJSON(w, http.StatusOK, map[string]any{"facilities": dtos})
	}
}

func toDTO(f facility.Facility) facilityDTO {
	return facilityDTO{
		ID:     f.ID,
		Name:   f.Name,
		Region: string(f.Region),
		Town:   f.Town,
		Beds:   f.Beds,
		Status: string(f.Status),
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
