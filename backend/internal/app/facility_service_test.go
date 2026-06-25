package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/payer"
	"github.com/xcreativs/gigmann/internal/core/severity"
)

type fakeRepo struct {
	items []facility.Facility
	err   error
}

func (f fakeRepo) List(context.Context) ([]facility.Facility, error) {
	return f.items, f.err
}

func mustFacility(t *testing.T) facility.Facility {
	t.Helper()
	mix, _ := payer.New(70, 25, 5)
	f, err := facility.New(facility.Params{
		ID: "f1", Name: "Kasoa", Region: "Central", Town: "Kasoa",
		Type: "OPD", Beds: 40, Lifecycle: facility.LifecycleActive, Health: severity.Good,
		ManagerName: "Ama Owusu", PayerMix: mix,
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	return f
}

func TestFacilityServiceList(t *testing.T) {
	svc := app.NewFacilityService(fakeRepo{items: []facility.Facility{mustFacility(t)}})

	got, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 1 || got[0].ID != "f1" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestFacilityServiceListError(t *testing.T) {
	svc := app.NewFacilityService(fakeRepo{err: errors.New("boom")})
	if _, err := svc.List(context.Background()); err == nil {
		t.Fatal("expected error to propagate")
	}
}
