// Package bootstrap is the composition root: it wires outbound/inbound adapters
// to the application use cases and runs the HTTP server with graceful shutdown.
// It holds only wiring (no business logic), so it carries no unit tests.
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/xcreativs/gigmann/internal/adapters/inbound/httpapi"
	"github.com/xcreativs/gigmann/internal/adapters/inbound/realtime"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/anthropic"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/audit"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/localembedder"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/localnarrator"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/passwordhash"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/postgres"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/token"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/voyage"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/webpush"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/config"
	signalengine "github.com/xcreativs/gigmann/internal/core/signal"
	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/observability"
	"github.com/xcreativs/gigmann/internal/ports"
	"github.com/xcreativs/gigmann/internal/seed"
	"github.com/xcreativs/gigmann/migrations"
)

const (
	shutdownTimeout = 10 * time.Second
	readTimeout     = 10 * time.Second
	writeTimeout    = 45 * time.Second // headroom for synchronous LLM calls (Ask)
	demoSeed        = 42
	briefTopN       = 5
	briefCacheTTL   = 10 * time.Minute
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

// repos bundles the persistence ports selected at wiring time.
type repos struct {
	facilities ports.FacilityRepository
	metrics    ports.MetricsRepository
	embeddings ports.FacilityEmbeddingRepository
	users      ports.UserRepository
	refresh    ports.RefreshTokenStore
	resets     ports.PasswordResetTokenStore
	approvals  ports.ApprovalRepository
	tasks      ports.TaskRepository
	ready      func(context.Context) error
}

// Run loads configuration, wires dependencies, and serves HTTP until interrupted.
func Run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	shutdownTracing, err := observability.SetupTracing(context.Background(), "gigmann-api", cfg.AppEnv)
	if err != nil {
		return err
	}
	defer func() { _ = shutdownTracing(context.Background()) }()

	flushSentry, err := observability.SetupErrorTracking(cfg.SentryDSN, cfg.AppEnv)
	if err != nil {
		return err
	}
	defer flushSentry()

	handler, cleanup, err := newHandler(context.Background(), cfg, logger)
	if err != nil {
		return err
	}
	defer cleanup()

	srv := &http.Server{
		Addr:         cfg.HTTPAddr(),
		Handler:      otelhttp.NewHandler(handler, "gigmann-api"),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("api listening", "addr", cfg.HTTPAddr(), "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	select {
	case err := <-serverErr:
		return fmt.Errorf("bootstrap: http server failed: %w", err)
	case <-stop:
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}
	logger.Info("shutdown complete")
	return nil
}

// newHandler wires repositories, the narrator, and the brief pipeline by config.
func newHandler(ctx context.Context, cfg config.Config, logger *slog.Logger) (http.Handler, func(), error) {
	net := seed.Generate(demoSeed, time.Now(), seed.DefaultDays)

	hasher := passwordhash.New()
	accounts, err := demoAccounts(hasher)
	if err != nil {
		return nil, nil, err
	}

	r, cleanup, err := selectRepos(ctx, cfg, net, accounts, logger)
	if err != nil {
		return nil, nil, err
	}

	// EnsureSeeded only populates an EMPTY database, so new demo identities would
	// never reach an already-seeded (e.g. production) store. Additively reconcile
	// them here on every boot — missing accounts are inserted, existing ones left
	// untouched. On a fresh in-memory store every account is already present, so
	// this is a no-op there.
	if added, rerr := reconcileDemoAccounts(ctx, r.users, accounts); rerr != nil {
		cleanup()
		return nil, nil, fmt.Errorf("bootstrap: reconcile demo accounts: %w", rerr)
	} else if added > 0 {
		logger.Info("added missing demo accounts", "count", added)
	}

	engine := signalengine.Default(signalengine.DefaultThresholds())
	var (
		narrator ports.Narrator
		answerer ports.Answerer
	)
	if cfg.AnthropicAPIKey != "" && cfg.Flags.AINarration {
		n := anthropic.NewNarrator(cfg.AnthropicAPIKey, cfg.AnthropicModel)
		narrator, answerer = n, n
		logger.Info("using Claude narrator", "model", cfg.AnthropicModel)
	} else {
		n := localnarrator.New()
		narrator, answerer = n, n
		logger.Info("no ANTHROPIC_API_KEY set — using the deterministic local narrator")
	}

	// Embedder: Voyage when configured, else the deterministic local fallback.
	var embedder ports.Embedder
	if cfg.VoyageAPIKey != "" {
		embedder = voyage.NewEmbedder(cfg.VoyageAPIKey, cfg.VoyageModel)
		logger.Info("using Voyage embeddings", "model", cfg.VoyageModel)
	} else {
		embedder = localembedder.New()
		logger.Info("no VOYAGE_API_KEY set — using the deterministic local embedder")
	}
	// Best-effort first-run embedding of facilities: NL search degrades to empty
	// if this fails, but it never blocks startup.
	if !cfg.Flags.FacilitySearch {
		logger.Info("FEATURE_FACILITY_SEARCH disabled — NL facility search returns no matches")
	} else if embedded, eerr := app.SeedFacilityEmbeddings(ctx, embedder, r.embeddings, net.Facilities); eerr != nil {
		logger.Warn("facility embedding seed failed; NL facility search disabled until retried", "err", eerr)
	} else if embedded {
		logger.Info("seeded facility embeddings", "facilities", len(net.Facilities))
	}
	searchSvc := app.NewFacilitySearchService(embedder, r.embeddings, net.Facilities)
	preferencesSvc := app.NewPreferencesService(r.users)

	briefSvc := app.NewBriefService(engine, narrator, briefTopN)
	input := signalengine.Input{
		AsOf: time.Now().UTC(), Facilities: net.Facilities, Metrics: net.Metrics,
		Inventory: net.Inventory, Staff: net.Staff,
	}
	briefs := app.NewCachedBrief(app.NewStaticBrief(briefSvc, input), briefCacheTTL)
	hub := realtime.New()
	go func() { //nolint:contextcheck // Startup cache warm runs detached from any request.
		if _, err := briefs.Generate(context.Background()); err != nil {
			logger.Warn("brief cache warm failed", "err", err)
			return
		}
		logger.Info("brief cache warmed")
	}()
	askSvc := app.NewAskService(engine, answerer, input, 0)
	metricsSvc := app.NewMetricsService(r.metrics)
	detailSvc := app.NewFacilityDetailService(net.Facilities, net.Inventory, net.Staff, net.Alerts, net.Metrics)

	tokens := token.New([]byte(cfg.JWTSecret), accessTokenTTL)
	auditLog := audit.New(logger)
	authSvc := app.NewAuthService(r.users, hasher, tokens, r.refresh, r.resets, refreshTokenTTL, auditLog)
	approvalSvc := app.NewApprovalService(r.approvals, auditLog)
	taskSvc := app.NewTaskService(r.tasks)
	alertRepo := memory.NewAlertRepo(net.Alerts...)
	alertSvc := app.NewAlertService(alertRepo)
	draftSvc := app.NewDraftService(askSvc)

	// Web Push (GEC-69): a no-op unless VAPID keys are configured. The critical-
	// only sweep hangs off the same brief-refresh signal as the realtime hub.
	pushSender := webpush.New(cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDSubject)
	pushSvc := app.NewPushService(memory.NewPushRepo(), pushSender, alertRepo)
	briefs.SetNotifier(app.FanoutNotifier(hub, pushSvc))
	if pushSender.Enabled() {
		logger.Info("web push enabled (critical-only)")
	}

	return httpapi.NewRouter(httpapi.Deps{
		Facilities:     app.NewFacilityService(r.facilities, r.metrics),
		FacilityDetail: detailSvc,
		Metrics:        metricsSvc,
		Briefs:         briefs,
		Auth:           authSvc,
		Approvals:      approvalSvc,
		Tasks:          taskSvc,
		Ask:            askSvc,
		Alerts:         alertSvc,
		Drafts:         draftSvc,
		Search:         searchSvc,
		Preferences:    preferencesSvc,
		Push:           pushSvc,
		Tokens:         tokens,
		Logger:         logger,
		CORSOrigins:    cfg.CORSAllowedOrigins,
		Realtime:       hub.Handler(tokens, cfg.CORSAllowedOrigins),
		HSTS:           cfg.IsProduction(),
		TrustProxy:     cfg.TrustProxy,
		Ready:          r.ready,
	}), cleanup, nil
}

// selectRepos chooses Postgres-backed repositories when DATABASE_URL is set
// (running migrations first and seeding an empty database), and otherwise the
// in-memory repositories seeded from the synthetic network.
func selectRepos(
	ctx context.Context, cfg config.Config, net seed.Network, accounts []ports.Account, logger *slog.Logger,
) (repos, func(), error) {
	if cfg.DatabaseURL == "" {
		// In-memory is the intended free-tier demo posture, but running it outside
		// development means all state (users, tokens, MFA, approvals, tasks) is
		// ephemeral and single-instance — surface that loudly rather than silently.
		if cfg.AppEnv != config.EnvDevelopment {
			logger.Warn("DATABASE_URL is unset — running on EPHEMERAL in-memory repositories; state is lost on restart and cannot scale past one instance",
				"env", cfg.AppEnv, "facilities", len(net.Facilities))
		} else {
			logger.Info("using in-memory repositories seeded from synthetic network", "facilities", len(net.Facilities))
		}
		return repos{
			facilities: memory.NewFacilityRepo(net.Facilities...),
			metrics:    memory.NewMetricsRepo(net.Metrics...),
			embeddings: memory.NewFacilityEmbeddingRepo(),
			users:      memory.NewUserRepo(accounts...),
			refresh:    memory.NewRefreshStore(),
			resets:     memory.NewPasswordResetStore(),
			approvals:  memory.NewApprovalRepo(net.Approvals...),
			tasks:      memory.NewTaskRepo(net.Tasks...),
			ready:      nil,
		}, func() {}, nil
	}

	if err := postgres.Migrate(ctx, cfg.DatabaseURL, migrations.Files); err != nil {
		return repos{}, nil, err
	}
	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return repos{}, nil, err
	}

	seeded, err := postgres.EnsureSeeded(ctx, pool, net.Facilities, net.Metrics, net.Approvals, net.Tasks, accounts)
	if err != nil {
		pool.Close()
		return repos{}, nil, err
	}
	if seeded {
		logger.Info("seeded empty postgres database", "facilities", len(net.Facilities))
	} else {
		logger.Info("postgres already seeded; preserving persisted data")
	}

	logger.Info("using postgres repositories")
	return repos{
		facilities: postgres.NewFacilityRepo(pool),
		metrics:    postgres.NewMetricsRepo(pool),
		embeddings: postgres.NewFacilityEmbeddingRepo(pool),
		users:      postgres.NewUserRepo(pool),
		refresh:    postgres.NewRefreshRepo(pool),
		resets:     postgres.NewPasswordResetRepo(pool),
		approvals:  postgres.NewApprovalRepo(pool),
		tasks:      postgres.NewTaskRepo(pool),
		ready:      pool.Ping,
	}, pool.Close, nil
}

// demoAccounts builds the demo login identities across both roles. The password
// comes from DEMO_PASSWORD (a low-entropy dev default otherwise) and is shared by
// every account. Executives are network-wide (no facility); facility managers are
// scoped to a real facility in the synthetic network so facility-scoped
// authorization is demonstrable — each manager sees only their own facility. The
// manager names mirror the seeded facility managers so the login identity matches
// the data.
func demoAccounts(hasher ports.PasswordHasher) ([]ports.Account, error) {
	password := os.Getenv("DEMO_PASSWORD")
	if password == "" {
		password = "ahenfie-demo" //nolint:gosec // demo seed password for local dev only
	}
	hash, err := hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: hash demo password: %w", err)
	}

	// facility == "" means an executive (network-wide, no facility scope).
	specs := []struct {
		id, name, email, facility string
		role                      user.Role
	}{
		// Executives.
		{"u-sammy", "Sammy Adjei", "ceo@gigmann.health", "", user.RoleExecutive},
		{"u-efua", "Efua Sarpong", "coo@gigmann.health", "", user.RoleExecutive},
		// One facility manager per facility (names mirror the seeded managers).
		{"u-ama", "Ama Owusu", "kasoa.manager@gigmann.health", "kasoa", user.RoleFacilityManager},
		{"u-kwame", "Dr. Kwame Mensah", "assin-fosu.manager@gigmann.health", "assin-fosu", user.RoleFacilityManager},
		{"u-afia", "Dr. Afia Boahen", "asokwa.manager@gigmann.health", "asokwa", user.RoleFacilityManager},
		{"u-yaw", "Yaw Antwi", "adansi.manager@gigmann.health", "adansi", user.RoleFacilityManager},
		{"u-esi", "Esi Quaye", "takoradi.manager@gigmann.health", "takoradi", user.RoleFacilityManager},
		{"u-adjoa", "Mad. Adjoa Asare", "tafo.manager@gigmann.health", "tafo-maternity", user.RoleFacilityManager},
		{"u-mohammed", "Mohammed Iddrisu", "nima.manager@gigmann.health", "nima", user.RoleFacilityManager},
		{"u-selorm", "Dr. Selorm Agbeko", "ho.manager@gigmann.health", "ho-central", user.RoleFacilityManager},
		{"u-fuseini", "Fuseini Abdulai", "tamale.manager@gigmann.health", "tamale-north", user.RoleFacilityManager},
		{"u-araba", "Dr. Araba Eshun", "cape-coast.manager@gigmann.health", "cape-coast", user.RoleFacilityManager},
		{"u-kwabena", "Kwabena Osei", "sunyani.manager@gigmann.health", "sunyani", user.RoleFacilityManager},
		{"u-akosua", "Akosua Mensimah", "sekondi.manager@gigmann.health", "sekondi", user.RoleFacilityManager},
	}

	accounts := make([]ports.Account, 0, len(specs))
	for _, s := range specs {
		u, uerr := user.New(user.User{ID: s.id, Name: s.name, Role: s.role, FacilityID: s.facility})
		if uerr != nil {
			return nil, fmt.Errorf("bootstrap: demo account %q: %w", s.id, uerr)
		}
		accounts = append(accounts, ports.Account{User: u, Email: s.email, PasswordHash: hash})
	}
	return accounts, nil
}

// reconcileDemoAccounts inserts any demo account not already present (matched by
// email). It is additive and never overwrites an existing account, so a user who
// changed their password or enrolled MFA is left untouched — only genuinely
// missing identities are added. EnsureSeeded seeds accounts only into an empty
// database, so this is what lets new demo identities appear in an already-seeded
// (e.g. production) database on the next boot.
func reconcileDemoAccounts(ctx context.Context, users ports.UserRepository, accounts []ports.Account) (int, error) {
	added := 0
	for _, acct := range accounts {
		switch _, err := users.FindByEmail(ctx, acct.Email); {
		case err == nil:
			continue // already present — do not clobber
		case errors.Is(err, ports.ErrAccountNotFound):
			if serr := users.Save(ctx, acct); serr != nil {
				return added, fmt.Errorf("save %s: %w", acct.Email, serr)
			}
			added++
		default:
			return added, fmt.Errorf("lookup %s: %w", acct.Email, err)
		}
	}
	return added, nil
}
