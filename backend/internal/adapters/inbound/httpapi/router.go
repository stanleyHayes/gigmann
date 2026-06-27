// Package httpapi is the inbound HTTP adapter (Chi). Handlers implement the
// generated OpenAPI strict server interface (see openapi_gen.go) and stay thin:
// they translate between transport types and application use cases.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/alert"
	"github.com/xcreativs/gigmann/internal/core/approval"
	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/inventory"
	"github.com/xcreativs/gigmann/internal/core/kpi"
	"github.com/xcreativs/gigmann/internal/core/staff"
	"github.com/xcreativs/gigmann/internal/core/task"
	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/intel"
	"github.com/xcreativs/gigmann/internal/ports"
)

const (
	requestTimeout = 45 * time.Second // the Ask endpoint calls the LLM synchronously (~20s)
	bearerPrefix   = "Bearer "
	authRateLimit  = 10
	authRateWindow = time.Minute
	askRateLimit   = 20
	askRateWindow  = time.Minute
)

// authRateLimitedPaths are the brute-force-sensitive auth endpoints.
var authRateLimitedPaths = []string{"/api/v1/auth/login", "/api/v1/auth/refresh"}

// askRateLimitedPaths bound AI cost/abuse per principal (GEC-48).
var askRateLimitedPaths = []string{"/api/v1/ask"}

// Deps are the application use cases the HTTP layer delegates to.
type Deps struct {
	Facilities     *app.FacilityService
	FacilityDetail *app.FacilityDetailService
	Metrics        *app.MetricsService
	Briefs         ports.BriefGenerator
	Auth           *app.AuthService
	Approvals      *app.ApprovalService
	Tasks          *app.TaskService
	Ask            ports.QuestionAnswerer
	Alerts         *app.AlertService
	Search         *app.FacilitySearchService
	Preferences    *app.PreferencesService
	Tokens         ports.TokenService
	Logger         *slog.Logger
	CORSOrigins    []string
	HSTS           bool
	Ready          func(context.Context) error
}

// Server implements the generated StrictServerInterface, delegating to use cases.
type Server struct {
	facilities     *app.FacilityService
	facilityDetail *app.FacilityDetailService
	metrics        *app.MetricsService
	briefs         ports.BriefGenerator
	auth           *app.AuthService
	approvals      *app.ApprovalService
	tasks          *app.TaskService
	ask            ports.QuestionAnswerer
	alerts         *app.AlertService
	search         *app.FacilitySearchService
	preferences    *app.PreferencesService
}

// Compile-time guarantee that Server satisfies the generated contract.
var _ StrictServerInterface = (*Server)(nil)

// NewRouter builds the HTTP handler from the generated OpenAPI contract.
func NewRouter(d Deps) http.Handler {
	logger := d.Logger
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(sentryMiddleware())
	metrics, metricsReg := newMetrics()
	r.Use(requestLogger(logger))
	r.Use(metrics.middleware())
	r.Use(securityHeaders(d.HSTS))
	r.Use(corsMiddleware(d.CORSOrigins))
	r.Use(rateLimit(newRateLimiter(authRateLimit, authRateWindow), authRateLimitedPaths))
	r.Use(middleware.Timeout(requestTimeout))
	r.Use(authMiddleware(d.Tokens))
	r.Use(rateLimitPrincipal(newRateLimiter(askRateLimit, askRateWindow), askRateLimitedPaths))
	r.Get("/readyz", readyHandler(d.Ready))
	r.Handle("/metrics", metricsHandler(metricsReg))
	r.Get("/openapi.json", openAPIHandler(logger))
	r.Get("/docs", redocHandler())

	srv := &Server{facilities: d.Facilities, facilityDetail: d.FacilityDetail, metrics: d.Metrics, briefs: d.Briefs, auth: d.Auth, approvals: d.Approvals, tasks: d.Tasks, ask: d.Ask, alerts: d.Alerts, search: d.Search, preferences: d.Preferences}
	return HandlerFromMux(NewStrictHandler(srv, []StrictMiddlewareFunc{requireAuth()}), r)
}

type ctxKey int

const principalKey ctxKey = iota

// openAPIHandler serves the embedded OpenAPI spec as JSON (browsable API docs;
// importable into Postman/Insomnia/Swagger). Public, read-only.
func openAPIHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		spec, err := GetSwagger()
		if err != nil {
			logger.Error("load openapi spec", "err", err)
			http.Error(w, `{"error":"internal_error"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(spec); err != nil {
			logger.Error("encode openapi spec", "err", err)
		}
	}
}

// redocPage renders the spec with Redoc (loaded from CDN). It overrides the strict
// default CSP for this route only.
const redocPage = `<!DOCTYPE html><html><head><title>Gigmann API · Ahenfie</title>` +
	`<meta charset="utf-8"/><meta name="viewport" content="width=device-width,initial-scale=1"/>` +
	`<style>body{margin:0}</style></head><body><redoc spec-url="/openapi.json"></redoc>` +
	`<script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script></body></html>`

func redocHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self' https://cdn.redoc.ly 'unsafe-inline'; "+
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; "+
				"img-src 'self' data: https:; worker-src 'self' blob:; connect-src 'self'")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(redocPage))
	}
}

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
	"PostAuthLogin":   true,
	"PostAuthRefresh": true,
	"PostAuthLogout":  true,
	"GetHealthz":      true,
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

// GetFacility returns one facility's drill-down (inventory, staff, alerts).
func (s *Server) GetFacility(ctx context.Context, request GetFacilityRequestObject) (GetFacilityResponseObject, error) {
	d, err := s.facilityDetail.Detail(ctx, request.FacilityId)
	switch {
	case errors.Is(err, app.ErrFacilityNotFound):
		return GetFacility404JSONResponse{NotFoundJSONResponse{Error: "not_found"}}, nil
	case err != nil:
		return GetFacility500JSONResponse{InternalErrorJSONResponse{Error: "internal_error"}}, nil //nolint:nilerr
	}
	return GetFacility200JSONResponse(toAPIFacilityDetail(d)), nil
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

// SearchFacilities resolves a natural-language phrase to facilities via vector
// similarity (GEC-13). Read-only; never triggers a side effect.
func (s *Server) SearchFacilities(ctx context.Context, request SearchFacilitiesRequestObject) (SearchFacilitiesResponseObject, error) {
	limit := 0
	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}
	matches, err := s.search.Resolve(ctx, request.Params.Q, limit)
	if err != nil {
		return SearchFacilities500JSONResponse{InternalErrorJSONResponse{Error: "internal_error"}}, nil //nolint:nilerr // mapped to 500; detail not leaked
	}
	return SearchFacilities200JSONResponse(toAPIFacilitySearch(request.Params.Q, matches)), nil
}

func toAPIFacilitySearch(query string, matches []app.FacilityMatch) FacilitySearchResults {
	out := make([]FacilityMatch, 0, len(matches))
	for _, m := range matches {
		out = append(out, FacilityMatch{FacilityId: m.FacilityID, Name: m.Name, Score: m.Score})
	}
	return FacilitySearchResults{Query: query, Matches: out}
}

// ListAlerts returns the ranked, cursor-paginated attention feed (open alerts).
func (s *Server) ListAlerts(ctx context.Context, request ListAlertsRequestObject) (ListAlertsResponseObject, error) {
	cursor := ""
	if request.Params.Cursor != nil {
		cursor = *request.Params.Cursor
	}
	limit := 0
	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}
	items, next, err := s.alerts.Feed(ctx, cursor, limit)
	if err != nil {
		return ListAlerts500JSONResponse{InternalErrorJSONResponse{Error: "internal_error"}}, nil //nolint:nilerr // mapped to 500
	}
	feed := AlertFeed{Alerts: make([]AlertItem, 0, len(items))}
	for _, a := range items {
		feed.Alerts = append(feed.Alerts, toAPIAlertItem(a))
	}
	if next != "" {
		feed.NextCursor = &next
	}
	return ListAlerts200JSONResponse(feed), nil
}

// UpdateAlertStatus dismisses or resolves an alert (explicit, user-initiated).
func (s *Server) UpdateAlertStatus(ctx context.Context, request UpdateAlertStatusRequestObject) (UpdateAlertStatusResponseObject, error) {
	if request.Body == nil {
		return UpdateAlertStatus400JSONResponse{BadRequestJSONResponse{Error: "bad_request"}}, nil
	}
	updated, err := s.alerts.UpdateStatus(ctx, request.AlertId, alert.Status(request.Body.Status))
	switch {
	case errors.Is(err, ports.ErrAlertNotFound):
		return UpdateAlertStatus404JSONResponse{NotFoundJSONResponse{Error: "not_found"}}, nil
	case errors.Is(err, app.ErrInvalidAlertStatus):
		return UpdateAlertStatus400JSONResponse{BadRequestJSONResponse{Error: "bad_request"}}, nil
	case errors.Is(err, alert.ErrAlreadyTerminal):
		return UpdateAlertStatus409JSONResponse{ConflictJSONResponse{Error: "already_decided"}}, nil
	case err != nil:
		return UpdateAlertStatus500JSONResponse{InternalErrorJSONResponse{Error: "internal_error"}}, nil //nolint:nilerr // mapped to 500
	}
	return UpdateAlertStatus200JSONResponse(toAPIAlertItem(updated)), nil
}

// GetMePreferences returns the current user's personalisation preferences.
func (s *Server) GetMePreferences(ctx context.Context, _ GetMePreferencesRequestObject) (GetMePreferencesResponseObject, error) {
	p, ok := principalFrom(ctx)
	if !ok {
		return GetMePreferences401JSONResponse{UnauthorizedJSONResponse{Error: "unauthorized"}}, nil
	}
	prefs, err := s.preferences.Get(ctx, p.UserID)
	if err != nil {
		return GetMePreferences500JSONResponse{InternalErrorJSONResponse{Error: "internal_error"}}, nil //nolint:nilerr // mapped to 500
	}
	return GetMePreferences200JSONResponse(toAPIPreferences(prefs)), nil
}

// UpdateMePreferences replaces the current user's preferences (sanitised in the app layer).
func (s *Server) UpdateMePreferences(ctx context.Context, request UpdateMePreferencesRequestObject) (UpdateMePreferencesResponseObject, error) {
	p, ok := principalFrom(ctx)
	if !ok {
		return UpdateMePreferences401JSONResponse{UnauthorizedJSONResponse{Error: "unauthorized"}}, nil
	}
	var body Preferences
	if request.Body != nil {
		body = *request.Body
	}
	updated, err := s.preferences.Update(ctx, p.UserID, fromAPIPreferences(body))
	if err != nil {
		return UpdateMePreferences500JSONResponse{InternalErrorJSONResponse{Error: "internal_error"}}, nil //nolint:nilerr // mapped to 500
	}
	return UpdateMePreferences200JSONResponse(toAPIPreferences(updated)), nil
}

func toAPIPreferences(p user.Preferences) Preferences {
	watched := p.WatchedMetrics
	if watched == nil {
		watched = []string{}
	}
	thresholds := p.Thresholds
	if thresholds == nil {
		thresholds = map[string]float64{}
	}
	return Preferences{WatchedMetrics: watched, Thresholds: thresholds}
}

func fromAPIPreferences(in Preferences) user.Preferences {
	return user.Preferences{WatchedMetrics: in.WatchedMetrics, Thresholds: in.Thresholds}
}

// ListApprovals returns the approvals routed to the executive.
func (s *Server) ListApprovals(ctx context.Context, _ ListApprovalsRequestObject) (ListApprovalsResponseObject, error) {
	items, err := s.approvals.List(ctx)
	if err != nil {
		return ListApprovals500JSONResponse{InternalErrorJSONResponse{Error: "internal_error"}}, nil //nolint:nilerr
	}
	out := make([]Approval, 0, len(items))
	for _, a := range items {
		out = append(out, toAPIApproval(a))
	}
	return ListApprovals200JSONResponse{Approvals: out}, nil
}

// DecideApproval records an explicit approve/decline decision (executive only).
func (s *Server) DecideApproval(ctx context.Context, request DecideApprovalRequestObject) (DecideApprovalResponseObject, error) {
	if request.Body == nil {
		return DecideApproval404JSONResponse{NotFoundJSONResponse{Error: "not_found"}}, nil
	}
	p, _ := principalFrom(ctx)
	note := ""
	if request.Body.Note != nil {
		note = *request.Body.Note
	}
	a, err := s.approvals.Decide(ctx, p, request.ApprovalId, string(request.Body.Decision) == "approve", note, time.Now())
	switch {
	case errors.Is(err, app.ErrForbidden):
		return DecideApproval403JSONResponse{ForbiddenJSONResponse{Error: "forbidden"}}, nil
	case errors.Is(err, ports.ErrApprovalNotFound):
		return DecideApproval404JSONResponse{NotFoundJSONResponse{Error: "not_found"}}, nil
	case errors.Is(err, approval.ErrAlreadyDecided):
		return DecideApproval409JSONResponse{ConflictJSONResponse{Error: "already_decided"}}, nil
	case err != nil:
		return DecideApproval500JSONResponse{InternalErrorJSONResponse{Error: "internal_error"}}, nil //nolint:nilerr
	}
	return DecideApproval200JSONResponse(toAPIApproval(a)), nil
}

// ListTasks returns the executive's "My Day" tasks.
func (s *Server) ListTasks(ctx context.Context, _ ListTasksRequestObject) (ListTasksResponseObject, error) {
	items, err := s.tasks.List(ctx)
	if err != nil {
		return ListTasks500JSONResponse{InternalErrorJSONResponse{Error: "internal_error"}}, nil //nolint:nilerr
	}
	out := make([]Task, 0, len(items))
	for _, t := range items {
		out = append(out, toAPITask(t))
	}
	return ListTasks200JSONResponse{Tasks: out}, nil
}

// UpdateTaskStatus moves a task to a new status.
func (s *Server) UpdateTaskStatus(ctx context.Context, request UpdateTaskStatusRequestObject) (UpdateTaskStatusResponseObject, error) {
	if request.Body == nil {
		return UpdateTaskStatus404JSONResponse{NotFoundJSONResponse{Error: "not_found"}}, nil
	}
	t, err := s.tasks.UpdateStatus(ctx, request.TaskId, task.Status(request.Body.Status))
	switch {
	case errors.Is(err, ports.ErrTaskNotFound):
		return UpdateTaskStatus404JSONResponse{NotFoundJSONResponse{Error: "not_found"}}, nil
	case err != nil:
		return UpdateTaskStatus500JSONResponse{InternalErrorJSONResponse{Error: "internal_error"}}, nil //nolint:nilerr
	}
	return UpdateTaskStatus200JSONResponse(toAPITask(t)), nil
}

// PostAsk answers a natural-language question grounded in the network context.
func (s *Server) PostAsk(ctx context.Context, request PostAskRequestObject) (PostAskResponseObject, error) {
	if request.Body == nil {
		return PostAsk500JSONResponse{InternalErrorJSONResponse{Error: "bad_request"}}, nil
	}
	a, err := s.ask.Answer(ctx, request.Body.Question)
	if err != nil {
		return PostAsk500JSONResponse{InternalErrorJSONResponse{Error: "internal_error"}}, nil //nolint:nilerr
	}
	return PostAsk200JSONResponse(toAPIAnswer(a)), nil
}

// PostAuthLogin exchanges email/password for a signed access token.
func (s *Server) PostAuthLogin(ctx context.Context, request PostAuthLoginRequestObject) (PostAuthLoginResponseObject, error) {
	if request.Body == nil {
		return PostAuthLogin401JSONResponse{UnauthorizedJSONResponse{Error: "invalid_credentials"}}, nil
	}
	code := ""
	if request.Body.Code != nil {
		code = *request.Body.Code
	}
	access, refresh, p, err := s.auth.Login(ctx, request.Body.Email, request.Body.Password, code)
	switch {
	case errors.Is(err, app.ErrMFARequired):
		return PostAuthLogin401JSONResponse{UnauthorizedJSONResponse{Error: "mfa_required"}}, nil
	case err != nil:
		return PostAuthLogin401JSONResponse{UnauthorizedJSONResponse{Error: "invalid_credentials"}}, nil //nolint:nilerr
	}
	return PostAuthLogin200JSONResponse(AuthSession{Token: access, RefreshToken: refresh, User: toAPIAuthUser(p)}), nil
}

// GetAuthMe returns the current authenticated user (set by authMiddleware).
func (s *Server) GetAuthMe(ctx context.Context, _ GetAuthMeRequestObject) (GetAuthMeResponseObject, error) {
	p, ok := principalFrom(ctx)
	if !ok {
		return GetAuthMe401JSONResponse{UnauthorizedJSONResponse{Error: "unauthorized"}}, nil
	}
	return GetAuthMe200JSONResponse(toAPIAuthUser(p)), nil
}

// PostAuthMfaEnroll mints a TOTP secret + otpauth URI for the current user.
func (s *Server) PostAuthMfaEnroll(ctx context.Context, _ PostAuthMfaEnrollRequestObject) (PostAuthMfaEnrollResponseObject, error) {
	p, ok := principalFrom(ctx)
	if !ok {
		return PostAuthMfaEnroll401JSONResponse{UnauthorizedJSONResponse{Error: "unauthorized"}}, nil
	}
	secret, uri, err := s.auth.BeginMFAEnrollment(ctx, p)
	if err != nil {
		return PostAuthMfaEnroll500JSONResponse{InternalErrorJSONResponse{Error: "internal_error"}}, nil //nolint:nilerr
	}
	return PostAuthMfaEnroll200JSONResponse(MfaEnrollment{Secret: secret, OtpauthUri: uri}), nil
}

// PostAuthMfaConfirm activates MFA after the user proves a valid code.
func (s *Server) PostAuthMfaConfirm(ctx context.Context, request PostAuthMfaConfirmRequestObject) (PostAuthMfaConfirmResponseObject, error) {
	p, ok := principalFrom(ctx)
	if !ok {
		return PostAuthMfaConfirm401JSONResponse{UnauthorizedJSONResponse{Error: "unauthorized"}}, nil
	}
	if request.Body == nil {
		return PostAuthMfaConfirm400JSONResponse{BadRequestJSONResponse{Error: "bad_request"}}, nil
	}
	err := s.auth.ConfirmMFAEnrollment(ctx, p, request.Body.Secret, request.Body.Code)
	switch {
	case errors.Is(err, app.ErrInvalidMFACode):
		return PostAuthMfaConfirm400JSONResponse{BadRequestJSONResponse{Error: "invalid_code"}}, nil
	case err != nil:
		return PostAuthMfaConfirm500JSONResponse{InternalErrorJSONResponse{Error: "internal_error"}}, nil //nolint:nilerr
	}
	return PostAuthMfaConfirm204Response{}, nil
}

// PostAuthRefresh rotates a refresh token into a fresh access + refresh pair.
func (s *Server) PostAuthRefresh(ctx context.Context, request PostAuthRefreshRequestObject) (PostAuthRefreshResponseObject, error) {
	if request.Body == nil {
		return PostAuthRefresh401JSONResponse{UnauthorizedJSONResponse{Error: "invalid_token"}}, nil
	}
	access, refresh, p, err := s.auth.Refresh(ctx, request.Body.RefreshToken)
	if err != nil {
		return PostAuthRefresh401JSONResponse{UnauthorizedJSONResponse{Error: "invalid_token"}}, nil //nolint:nilerr
	}
	return PostAuthRefresh200JSONResponse(AuthSession{Token: access, RefreshToken: refresh, User: toAPIAuthUser(p)}), nil
}

// PostAuthLogout revokes a refresh token (best-effort) and returns 204.
func (s *Server) PostAuthLogout(ctx context.Context, request PostAuthLogoutRequestObject) (PostAuthLogoutResponseObject, error) {
	if request.Body != nil {
		_ = s.auth.Logout(ctx, request.Body.RefreshToken)
	}
	return PostAuthLogout204Response{}, nil
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

func toAPIFacilityDetail(d app.FacilityDetail) FacilityDetail {
	inv := make([]InventoryItem, 0, len(d.Inventory))
	for _, it := range d.Inventory {
		inv = append(inv, toAPIInventoryItem(it))
	}
	stf := make([]StaffMember, 0, len(d.Staff))
	for _, m := range d.Staff {
		stf = append(stf, toAPIStaffMember(m))
	}
	alerts := make([]AlertItem, 0, len(d.Alerts))
	for _, a := range d.Alerts {
		alerts = append(alerts, toAPIAlertItem(a))
	}
	return FacilityDetail{Facility: toAPIFacility(d.Facility), Inventory: inv, Staff: stf, Alerts: alerts}
}

func toAPIInventoryItem(it inventory.Item) InventoryItem {
	out := InventoryItem{
		Id:               it.ID,
		Name:             it.Name,
		Category:         it.Category,
		StockLevel:       int32(it.StockLevel), //nolint:gosec // small non-negative stock count
		StockoutImminent: it.StockOutImminent(),
	}
	if days, ok := it.DaysOfStock(); ok {
		out.DaysOfStock = &days
	}
	return out
}

func toAPIStaffMember(m staff.Member) StaffMember {
	out := StaffMember{
		Id:            m.ID,
		Name:          m.Name,
		Role:          m.Role,
		Status:        m.Status,
		AttritionRisk: m.AttritionRisk,
	}
	if !m.LicenceExpiry.IsZero() {
		expiry := openapi_types.Date{Time: m.LicenceExpiry}
		out.LicenceExpiry = &expiry
	}
	return out
}

func toAPIAlertItem(a alert.Alert) AlertItem {
	out := AlertItem{
		Id:       a.ID,
		Type:     a.Type,
		Severity: FacilityStatus(a.Severity),
		Title:    a.Title,
		Status:   string(a.Status),
	}
	if a.FacilityID != "" {
		fid := a.FacilityID
		out.FacilityId = &fid
	}
	if a.Detail != "" {
		detail := a.Detail
		out.Detail = &detail
	}
	return out
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

func toAPIApproval(a approval.Approval) Approval {
	out := Approval{
		Id:            a.ID,
		Type:          ApprovalType(a.Type),
		FacilityId:    a.FacilityID,
		AmountPesewas: a.Amount.Pesewas(),
		Title:         a.Title,
		RequestedBy:   a.RequestedBy,
		Status:        ApprovalStatus(a.Status),
		CreatedAt:     a.CreatedAt,
	}
	if a.Context != "" {
		ctxText := a.Context
		out.Context = &ctxText
	}
	if !a.DecidedAt.IsZero() {
		decidedAt := a.DecidedAt
		out.DecidedAt = &decidedAt
	}
	if a.DecisionNote != "" {
		note := a.DecisionNote
		out.DecisionNote = &note
	}
	return out
}

func toAPITask(t task.Task) Task {
	out := Task{
		Id:        t.ID,
		Title:     t.Title,
		Priority:  TaskPriority(t.Priority),
		Status:    TaskStatus(t.Status),
		Source:    TaskSource(t.Source),
		CreatedAt: t.CreatedAt,
	}
	if t.Detail != "" {
		detail := t.Detail
		out.Detail = &detail
	}
	if t.FacilityID != "" {
		facilityID := t.FacilityID
		out.FacilityId = &facilityID
	}
	if t.AssignedTo != "" {
		assignedTo := t.AssignedTo
		out.AssignedTo = &assignedTo
	}
	if !t.DueDate.IsZero() {
		dueDate := openapi_types.Date{Time: t.DueDate}
		out.DueDate = &dueDate
	}
	return out
}

func toAPIAnswer(a intel.Answer) Answer {
	out := Answer{Text: a.Text}
	if len(a.Citations) > 0 {
		citations := a.Citations
		out.Citations = &citations
	}
	return out
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
