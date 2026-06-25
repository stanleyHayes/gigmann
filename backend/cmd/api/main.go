// Command api is the composition root for the Gigmann Executive Cockpit API.
// It only wires dependencies together; business logic lives in internal/.
package main

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

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config load failed", "err", err)
		os.Exit(1)
	}

	// Skeleton: seed two facilities in memory. Replaced by the Postgres
	// adapter + seed service in Epic E1 (GEC-14, GEC-15).
	repo := memory.NewFacilityRepo(mustSeed()...)
	facilitySvc := app.NewFacilityService(repo)
	handler := httpapi.NewRouter(facilitySvc)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr(),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		logger.Info("api listening", "addr", cfg.HTTPAddr(), "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
	}
	logger.Info("shutdown complete")
}

func mustSeed() []facility.Facility {
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
