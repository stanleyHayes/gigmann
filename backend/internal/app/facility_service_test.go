package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/facility"
)

type fakeRepo struct {
	items []facility.Facility
	err   error
}

func (f fakeRepo) List(context.Context) ([]facility.Facility, error) {
	return f.items, f.err
}

func TestFacilityServiceList(t *testing.T) {
	f, err := facility.New("f1", "Kasoa", "Central", "Kasoa", 40, facility.StatusGood)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	svc := app.NewFacilityService(fakeRepo{items: []facility.Facility{f}})

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
