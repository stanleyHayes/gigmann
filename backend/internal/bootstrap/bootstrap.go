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
	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/config"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/payer"
	"github.com/xcreativs/gigmann/internal/core/severity"
)

const shutdownTimeout = 10 * time.Second

// Run loads configuration, wires dependencies, and serves HTTP until interrupted.
func Run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:         cfg.HTTPAddr(),
		Handler:      newHandler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
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

// newHandler wires the in-memory repository (replaced by Postgres in E1) into
// the facility use case and the HTTP router.
func newHandler() http.Handler {
	repo := memory.NewFacilityRepo(seedFacilities()...)
	return httpapi.NewRouter(app.NewFacilityService(repo))
}

// seedFacilities returns a couple of skeleton facilities (replaced by the
// synthetic-network generator in GEC-15).
func seedFacilities() []facility.Facility {
	mk := func(p facility.Params) facility.Facility {
		f, err := facility.New(p)
		if err != nil {
			panic(err)
		}
		return f
	}
	mixFosu, _ := payer.New(65, 25, 10)
	mixTafo, _ := payer.New(80, 18, 2)
	return []facility.Facility{
		mk(facility.Params{
			ID: "assin-fosu", Name: "Assin Fosu Specialist Hospital", Region: "Central", Town: "Assin Fosu",
			Type: "Specialist", Beds: 60, Lifecycle: facility.LifecycleFlagship, Health: severity.Good,
			ManagerName: "Dr. Mensah", PayerMix: mixFosu,
		}),
		mk(facility.Params{
			ID: "tafo-maternity", Name: "Tafo Maternity & Child Health", Region: "Ashanti", Town: "Old Tafo",
			Type: "Maternity", Beds: 25, Lifecycle: facility.LifecycleActive, Health: severity.Critical,
			ManagerName: "Mad. Adjoa", PayerMix: mixTafo,
		}),
	}
}
