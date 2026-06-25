// Package bootstrap is the composition root: it wires outbound/inbound adapters
// to the application use cases and runs the HTTP server with graceful shutdown.
// It holds only wiring (no business logic), so it carries no unit tests.
package bootstrap

import (
	"context"
	"errors"
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
	"github.com/xcreativs/gigmann/internal/adapters/outbound/postgres"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/config"
	signalengine "github.com/xcreativs/gigmann/internal/core/signal"
	"github.com/xcreativs/gigmann/internal/ports"
	"github.com/xcreativs/gigmann/internal/seed"
)

const (
	shutdownTimeout = 10 * time.Second
	readTimeout     = 10 * time.Second
	writeTimeout    = 15 * time.Second
	demoSeed        = 42
	briefTopN       = 5
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

	var narrator ports.Narrator = localnarrator.New()
	if cfg.AnthropicAPIKey != "" {
		logger.Info("using Claude narrator", "model", cfg.AnthropicModel)
		narrator = anthropic.NewNarrator(cfg.AnthropicAPIKey, cfg.AnthropicModel)
	} else {
		logger.Info("no ANTHROPIC_API_KEY set — using the deterministic local narrator")
	}

	briefSvc := app.NewBriefService(signalengine.Default(signalengine.DefaultThresholds()), narrator, briefTopN)
	input := signalengine.Input{
		AsOf: time.Now().UTC(), Facilities: net.Facilities, Metrics: net.Metrics,
		Inventory: net.Inventory, Staff: net.Staff,
	}
	briefs := app.NewStaticBrief(briefSvc, input)

	return httpapi.NewRouter(app.NewFacilityService(facRepo), briefs), cleanup, nil
}
