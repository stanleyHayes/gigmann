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

	"github.com/xcreativs/gigmann/internal/adapters/inbound/httpapi"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/anthropic"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/localnarrator"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/passwordhash"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/postgres"
	"github.com/xcreativs/gigmann/internal/adapters/outbound/token"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/config"
	signalengine "github.com/xcreativs/gigmann/internal/core/signal"
	"github.com/xcreativs/gigmann/internal/core/user"
	"github.com/xcreativs/gigmann/internal/ports"
	"github.com/xcreativs/gigmann/internal/seed"
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

// Run loads configuration, wires dependencies, and serves HTTP until interrupted.
func Run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	handler, cleanup, err := newHandler(context.Background(), cfg, logger)
	if err != nil {
		return err
	}
	defer cleanup()

	srv := &http.Server{
		Addr:         cfg.HTTPAddr(),
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	go func() {
		logger.Info("api listening", "addr", cfg.HTTPAddr(), "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

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

	facRepo := ports.FacilityRepository(memory.NewFacilityRepo(net.Facilities...))
	cleanup := func() {}
	if cfg.DatabaseURL != "" {
		pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, nil, err
		}
		logger.Info("using postgres repository")
		facRepo = postgres.NewFacilityRepo(pool)
		cleanup = pool.Close
	} else {
		logger.Info("using in-memory repository seeded from synthetic network", "facilities", len(net.Facilities))
	}

	engine := signalengine.Default(signalengine.DefaultThresholds())
	var (
		narrator ports.Narrator
		answerer ports.Answerer
	)
	if cfg.AnthropicAPIKey != "" {
		n := anthropic.NewNarrator(cfg.AnthropicAPIKey, cfg.AnthropicModel)
		narrator, answerer = n, n
		logger.Info("using Claude narrator", "model", cfg.AnthropicModel)
	} else {
		n := localnarrator.New()
		narrator, answerer = n, n
		logger.Info("no ANTHROPIC_API_KEY set — using the deterministic local narrator")
	}

	briefSvc := app.NewBriefService(engine, narrator, briefTopN)
	input := signalengine.Input{
		AsOf: time.Now().UTC(), Facilities: net.Facilities, Metrics: net.Metrics,
		Inventory: net.Inventory, Staff: net.Staff,
	}
	briefs := app.NewCachedBrief(app.NewStaticBrief(briefSvc, input), briefCacheTTL)
	go func() { //nolint:contextcheck // startup cache warm runs detached from the request
		if _, err := briefs.Generate(context.Background()); err != nil {
			logger.Warn("brief cache warm failed", "err", err)
			return
		}
		logger.Info("brief cache warmed")
	}()
	askSvc := app.NewAskService(engine, answerer, input, 0)
	metricsSvc := app.NewMetricsService(net.Metrics)

	hasher := passwordhash.New()
	accounts, err := demoAccounts(hasher)
	if err != nil {
		return nil, nil, err
	}
	tokens := token.New([]byte(cfg.JWTSecret), accessTokenTTL)
	authSvc := app.NewAuthService(memory.NewUserRepo(accounts...), hasher, tokens, memory.NewRefreshStore(), refreshTokenTTL)

	approvalSvc := app.NewApprovalService(memory.NewApprovalRepo(net.Approvals...))
	taskSvc := app.NewTaskService(memory.NewTaskRepo(net.Tasks...))

	return httpapi.NewRouter(httpapi.Deps{
		Facilities: app.NewFacilityService(facRepo),
		Metrics:    metricsSvc,
		Briefs:     briefs,
		Auth:       authSvc,
		Approvals:  approvalSvc,
		Tasks:      taskSvc,
		Ask:        askSvc,
		Tokens:     tokens,
	}), cleanup, nil
}

// demoAccounts seeds the in-memory user store for the demo. The password comes
// from DEMO_PASSWORD (a low-entropy dev default otherwise); real deployments use
// a database-backed user store.
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
	manager, err := user.New(user.User{ID: "u-ama", Name: "Ama Owusu", Role: user.RoleFacilityManager, FacilityID: "kasoa-polyclinic"})
	if err != nil {
		return nil, fmt.Errorf("bootstrap: manager account: %w", err)
	}
	return []ports.Account{
		{User: ceo, Email: "ceo@gigmann.health", PasswordHash: hash},
		{User: manager, Email: "kasoa.manager@gigmann.health", PasswordHash: hash},
	}, nil
}
