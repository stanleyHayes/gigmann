// Package httpapi is the inbound HTTP adapter (Chi). Handlers implement the
// generated OpenAPI strict server interface (see openapi_gen.go) and stay thin:
// they translate between transport types and application use cases.
package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/kpi"
	"github.com/xcreativs/gigmann/internal/ports"
)

const (
	requestTimeout = 15 * time.Second
	bearerPrefix   = "Bearer "
)

// Deps are the application use cases the HTTP layer delegates to.
type Deps struct {
	Facilities *app.FacilityService
	Metrics    *app.MetricsService
	Briefs     ports.BriefGenerator
	Auth       *app.AuthService
	Tokens     ports.TokenService
}

// Server implements the generated StrictServerInterface, delegating to use cases.
type Server struct {
	facilities *app.FacilityService
	metrics    *app.MetricsService
	briefs     ports.BriefGenerator
	auth       *app.AuthService
}

// Compile-time guarantee that Server satisfies the generated contract.
var _ StrictServerInterface = (*Server)(nil)

// NewRouter builds the HTTP handler from the generated OpenAPI contract.
func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(requestTimeout))
	r.Use(authMiddleware(d.Tokens))

	srv := &Server{facilities: d.Facilities, metrics: d.Metrics, briefs: d.Briefs, auth: d.Auth}
	return HandlerFromMux(NewStrictHandler(srv, []StrictMiddlewareFunc{requireAuth()}), r)
}

type ctxKey int

const principalKey ctxKey = iota

func withPrincipal(ctx context.Context, p auth.Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

func principalFrom(ctx context.Context) (auth.Principal, bool) {
	p, ok := ctx.Value(principalKey).(auth.Principal)
	return p, ok
}

// authMiddleware authenticates a Bearer token when present: a valid token puts
// the principal in the request context; an invalid one is rejected with 401.
// Requests without a token pass through unauthenticated (handlers enforce).
func authMiddleware(tokens ports.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				next.ServeHTTP(w, r)
				return
			}
			if !strings.HasPrefix(header, bearerPrefix) {
				writeUnauthorized(w)
				return
			}
			p, err := tokens.Verify(strings.TrimPrefix(header, bearerPrefix))
			if err != nil {
				writeUnauthorized(w)
				return
			}
			next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), p)))
		})
	}
}

// unauthorizedBody is the fixed 401 payload (matches the Error schema).
var unauthorizedBody = []byte(`{"error":"unauthorized"}`)

// publicOperations may be called without authentication.
var publicOperations = map[string]bool{
	"PostAuthLogin": true,
	"GetHealthz":    true,
}

// requireAuth rejects any non-public operation that lacks an authenticated
// principal (set by authMiddleware from a valid Bearer token). Enforced at the
// use-case boundary so every business endpoint is protected by default.
func requireAuth() StrictMiddlewareFunc {
	return func(next StrictHandlerFunc, operationID string) StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			if !publicOperations[operationID] {
				if _, ok := principalFrom(ctx); !ok {
					writeUnauthorized(w)
					return nil, nil
				}
			}
			return next(ctx, w, r, request)
		}
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write(unauthorizedBody)
}

// GetHealthz implements the liveness probe (also mounted at /readyz upstream).
func (s *Server) GetHealthz(_ context.Context, _ GetHealthzRequestObject) (GetHealthzResponseObject, error) {
	return GetHealthz200JSONResponse{Status: "ok"}, nil
}

// ListFacilities returns all facilities in the network.
func (s *Server) ListFacilities(ctx context.Context, _ ListFacilitiesRequestObject) (ListFacilitiesResponseObject, error) {
	items, err := s.facilities.List(ctx)
	if err != nil {
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

// GetMetrics returns the deterministic network KPIs and weekly trends.
func (s *Server) GetMetrics(ctx context.Context, _ GetMetricsRequestObject) (GetMetricsResponseObject, error) {
	n, err := s.metrics.Network(ctx)
	if err != nil {
		return GetMetrics500JSONResponse{InternalErrorJSONResponse{Error: "internal_error"}}, nil //nolint:nilerr
	}
	return GetMetrics200JSONResponse(toAPINetworkMetrics(n)), nil
}

// PostAuthLogin exchanges email/password for a signed access token.
func (s *Server) PostAuthLogin(ctx context.Context, request PostAuthLoginRequestObject) (PostAuthLoginResponseObject, error) {
	if request.Body == nil {
		return PostAuthLogin401JSONResponse{UnauthorizedJSONResponse{Error: "invalid_credentials"}}, nil
	}
	tok, p, err := s.auth.Login(ctx, request.Body.Email, request.Body.Password)
	if err != nil {
		return PostAuthLogin401JSONResponse{UnauthorizedJSONResponse{Error: "invalid_credentials"}}, nil //nolint:nilerr
	}
	return PostAuthLogin200JSONResponse(AuthSession{Token: tok, User: toAPIAuthUser(p)}), nil
}

// GetAuthMe returns the current authenticated user (set by authMiddleware).
func (s *Server) GetAuthMe(ctx context.Context, _ GetAuthMeRequestObject) (GetAuthMeResponseObject, error) {
	p, ok := principalFrom(ctx)
	if !ok {
		return GetAuthMe401JSONResponse{UnauthorizedJSONResponse{Error: "unauthorized"}}, nil
	}
	return GetAuthMe200JSONResponse(toAPIAuthUser(p)), nil
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

func toAPINetworkMetrics(n kpi.Network) NetworkMetrics {
	kpis := make([]Kpi, 0, len(n.KPIs))
	for _, k := range n.KPIs {
		points := make([]MetricPoint, 0, len(k.Series))
		for _, p := range k.Series {
			points = append(points, MetricPoint{Date: openapi_types.Date{Time: p.Date}, Value: p.Value})
		}
		kpis = append(kpis, Kpi{
			Key:            k.Key,
			Label:          k.Label,
			Unit:           KpiUnit(k.Unit),
			HigherIsBetter: k.HigherIsBetter,
			Current:        k.Current,
			Previous:       k.Previous,
			DeltaPct:       k.DeltaPct,
			Direction:      KpiDirection(k.Direction),
			Series:         points,
		})
	}
	return NetworkMetrics{AsOf: openapi_types.Date{Time: n.AsOf}, Kpis: kpis}
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

func toAPIAuthUser(p auth.Principal) AuthUser {
	out := AuthUser{Id: p.UserID, Name: p.Name, Role: AuthUserRole(p.Role)}
	if p.FacilityID != "" {
		fid := p.FacilityID
		out.FacilityId = &fid
	}
	return out
}
