// Package httpapi is the inbound HTTP adapter (Chi). Handlers implement the
// generated OpenAPI strict server interface (see openapi_gen.go) and stay thin:
// they translate between transport types and application use cases.
package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/ports"
)

// Server implements the generated StrictServerInterface, delegating to use cases.
type Server struct {
	facilities *app.FacilityService
	briefs     ports.BriefGenerator
}

// Compile-time guarantee that Server satisfies the generated contract.
var _ StrictServerInterface = (*Server)(nil)

// NewRouter builds the HTTP handler from the generated OpenAPI contract.
func NewRouter(facilities *app.FacilityService, briefs ports.BriefGenerator) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))

	srv := &Server{facilities: facilities, briefs: briefs}
	return HandlerFromMux(NewStrictHandler(srv, nil), r)
}

// GetHealthz implements the liveness probe (also mounted at /readyz upstream).
func (s *Server) GetHealthz(_ context.Context, _ GetHealthzRequestObject) (GetHealthzResponseObject, error) {
	return GetHealthz200JSONResponse{Status: "ok"}, nil
}

// ListFacilities returns all facilities in the network.
func (s *Server) ListFacilities(ctx context.Context, _ ListFacilitiesRequestObject) (ListFacilitiesResponseObject, error) {
	items, err := s.facilities.List(ctx)
	if err != nil {
		// The 500 is conveyed via the response object; the Go error return is
		// reserved for unexpected framework-level failures.
		return ListFacilities500JSONResponse{InternalErrorJSONResponse{Error: "internal_error"}}, nil //nolint:nilerr
	}
	out := make([]Facility, 0, len(items))
	for _, f := range items {
		out = append(out, toAPIFacility(f))
	}
	return ListFacilities200JSONResponse{Facilities: out}, nil
}

// GetBrief returns the AI-narrated Daily Brief over the current network.
func (s *Server) GetBrief(ctx context.Context, _ GetBriefRequestObject) (GetBriefResponseObject, error) {
	b, err := s.briefs.Generate(ctx)
	if err != nil {
		return GetBrief500JSONResponse{InternalErrorJSONResponse{Error: "internal_error"}}, nil //nolint:nilerr
	}
	return GetBrief200JSONResponse(toAPIBrief(b)), nil
}

func toAPIFacility(f facility.Facility) Facility {
	return Facility{
		Id:     f.ID,
		Name:   f.Name,
		Region: string(f.Region),
		Town:   f.Town,
		Beds:   int32(f.Beds), //nolint:gosec // beds is a small, non-negative bed count
		Status: FacilityStatus(f.Health),
	}
}

func toAPIBrief(b brief.Brief) Brief {
	generatedAt := b.GeneratedAt
	items := make([]BriefItem, 0, len(b.Items))
	for _, it := range b.Items {
		items = append(items, toAPIBriefItem(it))
	}
	return Brief{
		Id:          b.ID,
		Date:        openapi_types.Date{Time: b.Date},
		Prose:       b.Prose,
		Model:       b.Model,
		GeneratedAt: &generatedAt,
		Items:       items,
	}
}

func toAPIBriefItem(it brief.Item) BriefItem {
	out := BriefItem{
		Severity:   FacilityStatus(it.Severity),
		FacilityId: it.FacilityID,
		Headline:   it.Headline,
	}
	if it.Explanation != "" {
		explanation := it.Explanation
		out.Explanation = &explanation
	}
	if len(it.SuggestedActions) > 0 {
		actions := it.SuggestedActions
		out.SuggestedActions = &actions
	}
	return out
}
