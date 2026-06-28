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
	go func() { //nolint:contextcheck,gosec // G118: startup cache warm runs detached from any request
		if _, err := briefs.Generate(context.Background()); err != nil {
			logger.Warn("brief cache warm failed", "err", err)
			return
		}
		logger.Info("brief cache warmed")
	}()
	askSvc := app.NewAskService(engine, answerer, input, 0)
	metricsSvc := app.NewMetricsService(r.metrics)
	detailSvc := app.NewFacilityDetailService(net.Facilities, net.Inventory, net.Staff, net.Alerts)

	tokens := token.New([]byte(cfg.JWTSecret), accessTokenTTL)
	auditLog := audit.New(logger)
	authSvc := app.NewAuthService(r.users, hasher, tokens, r.refresh, refreshTokenTTL, auditLog)
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
		Facilities:     app.NewFacilityService(r.facilities),
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
		logger.Info("using in-memory repositories seeded from synthetic network", "facilities", len(net.Facilities))
		return repos{
			facilities: memory.NewFacilityRepo(net.Facilities...),
			metrics:    memory.NewMetricsRepo(net.Metrics...),
			embeddings: memory.NewFacilityEmbeddingRepo(),
			users:      memory.NewUserRepo(accounts...),
			refresh:    memory.NewRefreshStore(),
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
		approvals:  postgres.NewApprovalRepo(pool),
		tasks:      postgres.NewTaskRepo(pool),
		ready:      pool.Ping,
	}, pool.Close, nil
}

// demoAccounts seeds the user store for the demo. The password comes from
// DEMO_PASSWORD (a low-entropy dev default otherwise). The manager is scoped to a
// real facility in the synthetic network so facility-scoped authorization works.
func demoAccounts(hasher ports.PasswordHasher) ([]ports.Account, error) {
	password := os.Getenv("DEMO_PASSWORD")
	if password == "" {
		password = "ahenfie-demo" //nolint:gosec // demo seed password for local dev only
	}
	hash, err := hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: hash demo password: %w", err)
	}
	ceo, err := user.New(user.User{ID: "u-sammy", Name: "Sammy Adjei", Role: user.RoleExecutive})
	if err != nil {
		return nil, fmt.Errorf("bootstrap: ceo account: %w", err)
	}
	manager, err := user.New(user.User{ID: "u-ama", Name: "Ama Owusu", Role: user.RoleFacilityManager, FacilityID: "kasoa"})
	if err != nil {
		return nil, fmt.Errorf("bootstrap: manager account: %w", err)
	}
	return []ports.Account{
		{User: ceo, Email: "ceo@gigmann.health", PasswordHash: hash},
		{User: manager, Email: "kasoa.manager@gigmann.health", PasswordHash: hash},
	}, nil
}
