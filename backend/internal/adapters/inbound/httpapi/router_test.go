package httpapi_test

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/xcreativs/gigmann/internal/adapters/inbound/httpapi"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/audit"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/localnarrator"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/passwordhash"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/token"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/approval"
	"github.com/xcreativs/gigmann/internal/core/auth"
	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/mfa"
	"github.com/xcreativs/gigmann/internal/core/payer"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/core/signal"
	"github.com/xcreativs/gigmann/internal/core/task"
	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/ports"
	"github.com/xcreativs/gigmann/internal/ports/mocks"
	"github.com/xcreativs/gigmann/internal/seed"
)

const (
	testEmail    = "ceo@gigmann.health"
	testPassword = "demo-pass-1234"
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

// newTestRouter builds the full HTTP handler with a seeded auth stack (a known
// executive account) so handler tests can authenticate end to end.
func newTestRouter(t *testing.T, repo *mocks.MockFacilityRepository, briefs *mocks.MockBriefGenerator) http.Handler {
	t.Helper()
	hasher := passwordhash.New()
	hash, err := hasher.Hash(testPassword)
	require.NoError(t, err)
	u, err := user.New(user.User{ID: "u1", Name: "Sammy Adjei", Role: user.RoleExecutive})
	require.NoError(t, err)
	users := memory.NewUserRepo(ports.Account{User: u, Email: testEmail, PasswordHash: hash})
	tokens := token.New([]byte("test-secret"), time.Hour)
	metricsSvc := app.NewMetricsService(seed.Generate(7, time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC), 14).Metrics)
	auditLog := audit.New(slog.New(slog.DiscardHandler))
	approvalSvc := app.NewApprovalService(memory.NewApprovalRepo(approval.Approval{
		ID: "ap-test", Type: approval.TypeCapital, FacilityID: "kasoa", Title: "Test approval",
		RequestedBy: "Ama Owusu", Status: approval.StatusPending,
	}), auditLog)
	taskSvc := app.NewTaskService(memory.NewTaskRepo(task.Task{
		ID: "task-test", Title: "Test task", Priority: task.PriorityHigh, Status: task.StatusTodo, Source: task.SourceBrief,
	}))
	net := seed.Generate(7, time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC), 14)
	askSvc := app.NewAskService(signal.Default(signal.DefaultThresholds()), localnarrator.New(),
		signal.Input{AsOf: net.Metrics[0].Date, Facilities: net.Facilities, Metrics: net.Metrics, Inventory: net.Inventory, Staff: net.Staff}, 0)

	detailSvc := app.NewFacilityDetailService(net.Facilities, net.Inventory, net.Staff, net.Alerts)

	return httpapi.NewRouter(httpapi.Deps{
		Facilities:     app.NewFacilityService(repo),
		FacilityDetail: detailSvc,
		Metrics:        metricsSvc,
		Briefs:         briefs,
		Auth:           app.NewAuthService(users, hasher, tokens, memory.NewRefreshStore(), time.Hour, auditLog),
		Approvals:      approvalSvc,
		Tasks:          taskSvc,
		Ask:            askSvc,
		Tokens:         tokens,
		CORSOrigins:    []string{"http://localhost:5173"},
	})
}

func serve(t *testing.T, repo *mocks.MockFacilityRepository, briefs *mocks.MockBriefGenerator, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, nil)
	newTestRouter(t, repo, briefs).ServeHTTP(rec, req)
	return rec
}

// bearerToken issues a valid executive token signed with the test secret.
func bearerToken(t *testing.T) string {
	t.Helper()
	tok, err := token.New([]byte("test-secret"), time.Hour).Issue(
		auth.Principal{UserID: "u1", Name: "Sammy Adjei", Role: user.RoleExecutive})
	require.NoError(t, err)
	return tok
}

// serveAuth issues a request carrying a valid Bearer token (for protected endpoints).
func serveAuth(t *testing.T, repo *mocks.MockFacilityRepository, briefs *mocks.MockBriefGenerator, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, nil)
	req.Header.Set("Authorization", "Bearer "+bearerToken(t))
	newTestRouter(t, repo, briefs).ServeHTTP(rec, req)
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

	rec := serveAuth(t, repo, mocks.NewMockBriefGenerator(ctrl), http.MethodGet, "/api/v1/facilities")

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

	rec := serveAuth(t, repo, mocks.NewMockBriefGenerator(ctrl), http.MethodGet, "/api/v1/facilities")
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

	rec := serveAuth(t, mocks.NewMockFacilityRepository(ctrl), briefs, http.MethodGet, "/api/v1/brief")

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

	rec := serveAuth(t, mocks.NewMockFacilityRepository(ctrl), briefs, http.MethodGet, "/api/v1/brief")
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	rec := serveAuth(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl), http.MethodGet, "/api/v1/metrics")

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

func postJSON(t *testing.T, h http.Handler, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	return rec
}

func TestAuthLoginSuccessAndMe(t *testing.T) {
	ctrl := gomock.NewController(t)
	h := newTestRouter(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl))

	rec := postJSON(t, h, "/api/v1/auth/login", `{"email":"ceo@gigmann.health","password":"demo-pass-1234"}`)
	require.Equal(t, http.StatusOK, rec.Code)
	var session struct {
		Token        string         `json:"token"`
		RefreshToken string         `json:"refresh_token"`
		User         map[string]any `json:"user"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&session))
	require.NotEmpty(t, session.Token)
	require.NotEmpty(t, session.RefreshToken)
	assert.Equal(t, "executive", session.User["role"])

	meRec := httptest.NewRecorder()
	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+session.Token)
	h.ServeHTTP(meRec, meReq)
	require.Equal(t, http.StatusOK, meRec.Code)
	var me map[string]any
	require.NoError(t, json.NewDecoder(meRec.Body).Decode(&me))
	assert.Equal(t, "Sammy Adjei", me["name"])
}

func TestAuthLoginBadPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	h := newTestRouter(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl))
	rec := postJSON(t, h, "/api/v1/auth/login", `{"email":"ceo@gigmann.health","password":"wrong"}`)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMeRequiresToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	rec := serve(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl), http.MethodGet, "/api/v1/auth/me")
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMeRejectsBadToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	h := newTestRouter(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestProtectedEndpointRequiresAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	for _, target := range []string{"/api/v1/facilities", "/api/v1/brief", "/api/v1/metrics", "/api/v1/approvals", "/api/v1/tasks"} {
		rec := serve(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl), http.MethodGet, target)
		require.Equal(t, http.StatusUnauthorized, rec.Code, target)
	}
}

func loginForRefresh(t *testing.T, h http.Handler) string {
	t.Helper()
	rec := postJSON(t, h, "/api/v1/auth/login", `{"email":"ceo@gigmann.health","password":"demo-pass-1234"}`)
	require.Equal(t, http.StatusOK, rec.Code)
	var s struct {
		RefreshToken string `json:"refresh_token"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&s))
	require.NotEmpty(t, s.RefreshToken)
	return s.RefreshToken
}

func TestAuthRefreshRotates(t *testing.T) {
	ctrl := gomock.NewController(t)
	h := newTestRouter(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl))
	refreshToken := loginForRefresh(t, h)

	rec := postJSON(t, h, "/api/v1/auth/refresh", `{"refresh_token":"`+refreshToken+`"}`)
	require.Equal(t, http.StatusOK, rec.Code)
	var rotated struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&rotated))
	require.NotEmpty(t, rotated.Token)
	require.NotEmpty(t, rotated.RefreshToken)
	assert.NotEqual(t, refreshToken, rotated.RefreshToken, "refresh token must rotate")

	// the old refresh token is single-use: reusing it fails
	reuse := postJSON(t, h, "/api/v1/auth/refresh", `{"refresh_token":"`+refreshToken+`"}`)
	require.Equal(t, http.StatusUnauthorized, reuse.Code)
}

func TestAuthRefreshRejectsBadToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	h := newTestRouter(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl))
	rec := postJSON(t, h, "/api/v1/auth/refresh", `{"refresh_token":"never-issued"}`)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthLogoutRevokes(t *testing.T) {
	ctrl := gomock.NewController(t)
	h := newTestRouter(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl))
	refreshToken := loginForRefresh(t, h)

	rec := postJSON(t, h, "/api/v1/auth/logout", `{"refresh_token":"`+refreshToken+`"}`)
	require.Equal(t, http.StatusNoContent, rec.Code)

	// after logout the refresh token can no longer be rotated
	after := postJSON(t, h, "/api/v1/auth/refresh", `{"refresh_token":"`+refreshToken+`"}`)
	require.Equal(t, http.StatusUnauthorized, after.Code)
}

func postAuthJSON(t *testing.T, h http.Handler, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearerToken(t))
	h.ServeHTTP(rec, req)
	return rec
}

func TestListApprovals(t *testing.T) {
	ctrl := gomock.NewController(t)
	rec := serveAuth(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl), http.MethodGet, "/api/v1/approvals")
	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Approvals []map[string]any `json:"approvals"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Len(t, body.Approvals, 1)
	assert.Equal(t, "ap-test", body.Approvals[0]["id"])
}

func TestDecideApprovalApproves(t *testing.T) {
	ctrl := gomock.NewController(t)
	h := newTestRouter(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl))
	rec := postAuthJSON(t, h, "/api/v1/approvals/ap-test/decision", `{"decision":"approve","note":"Go ahead"}`)
	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, "approved", body["status"])
	assert.Equal(t, "Go ahead", body["decision_note"])
}

func TestDecideApprovalNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	h := newTestRouter(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl))
	rec := postAuthJSON(t, h, "/api/v1/approvals/missing/decision", `{"decision":"approve"}`)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDecideApprovalRequiresAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	h := newTestRouter(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl))
	rec := postJSON(t, h, "/api/v1/approvals/ap-test/decision", `{"decision":"approve"}`)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestListTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	rec := serveAuth(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl), http.MethodGet, "/api/v1/tasks")
	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Tasks []map[string]any `json:"tasks"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Len(t, body.Tasks, 1)
	assert.Equal(t, "task-test", body.Tasks[0]["id"])
}

func TestUpdateTaskStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	h := newTestRouter(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl))
	rec := postAuthJSON(t, h, "/api/v1/tasks/task-test/status", `{"status":"done"}`)
	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, "done", body["status"])
}

func TestUpdateTaskStatusNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	h := newTestRouter(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl))
	rec := postAuthJSON(t, h, "/api/v1/tasks/missing/status", `{"status":"done"}`)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAsk(t *testing.T) {
	ctrl := gomock.NewController(t)
	h := newTestRouter(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl))
	rec := postAuthJSON(t, h, "/api/v1/ask", `{"question":"What is happening across the network?"}`)
	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Text string `json:"text"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.NotEmpty(t, body.Text)
}

func TestAskRequiresAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	h := newTestRouter(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl))
	rec := postJSON(t, h, "/api/v1/ask", `{"question":"x"}`)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestReadyz(t *testing.T) {
	ctrl := gomock.NewController(t)
	rec := serve(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl), http.MethodGet, "/readyz")
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestGetFacilityDetail(t *testing.T) {
	ctrl := gomock.NewController(t)
	id := seed.Generate(7, time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC), 14).Facilities[0].ID
	rec := serveAuth(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl), http.MethodGet, "/api/v1/facilities/"+id)
	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Facility map[string]any   `json:"facility"`
		Staff    []map[string]any `json:"staff"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, id, body.Facility["id"])
}

func TestGetFacilityDetailNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	rec := serveAuth(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl), http.MethodGet, "/api/v1/facilities/ghost-facility")
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestMFAEnrollAndStepUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	h := newTestRouter(t, mocks.NewMockFacilityRepository(ctrl), mocks.NewMockBriefGenerator(ctrl))

	// initial login (no MFA enrolled yet)
	rec := postJSON(t, h, "/api/v1/auth/login", `{"email":"ceo@gigmann.health","password":"demo-pass-1234"}`)
	require.Equal(t, http.StatusOK, rec.Code)
	var sess struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&sess))

	authed := func(method, target, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		var r *http.Request
		if body == "" {
			r = httptest.NewRequest(method, target, nil)
		} else {
			r = httptest.NewRequest(method, target, strings.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
		}
		r.Header.Set("Authorization", "Bearer "+sess.Token)
		h.ServeHTTP(w, r)
		return w
	}

	// enroll → get a secret
	enroll := authed(http.MethodPost, "/api/v1/auth/mfa/enroll", "")
	require.Equal(t, http.StatusOK, enroll.Code)
	var enr struct {
		Secret string `json:"secret"`
	}
	require.NoError(t, json.NewDecoder(enroll.Body).Decode(&enr))
	require.NotEmpty(t, enr.Secret)

	// confirm with a valid code → activates MFA
	code, err := mfa.Code(enr.Secret, time.Now())
	require.NoError(t, err)
	confirm := authed(http.MethodPost, "/api/v1/auth/mfa/confirm", `{"secret":"`+enr.Secret+`","code":"`+code+`"}`)
	require.Equal(t, http.StatusNoContent, confirm.Code)

	// login without a code now requires MFA
	noCode := postJSON(t, h, "/api/v1/auth/login", `{"email":"ceo@gigmann.health","password":"demo-pass-1234"}`)
	require.Equal(t, http.StatusUnauthorized, noCode.Code)

	// login with a valid code succeeds
	code2, err := mfa.Code(enr.Secret, time.Now())
	require.NoError(t, err)
	withCode := postJSON(t, h, "/api/v1/auth/login", `{"email":"ceo@gigmann.health","password":"demo-pass-1234","code":"`+code2+`"}`)
	require.Equal(t, http.StatusOK, withCode.Code)
}
