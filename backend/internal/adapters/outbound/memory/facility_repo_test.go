package memory_test

import (
	"context"
	"testing"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/payer"
	"github.com/xcreativs/gigmann/internal/core/severity"
)

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

func TestFacilityRepoList(t *testing.T) {
	repo := memory.NewFacilityRepo(mustFacility(t))
	got, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 facility, got %d", len(got))
	}

	// Returned slice must be a copy: mutating it must not affect the repo.
	got[0].Name = "mutated"
	again, _ := repo.List(context.Background())
	if again[0].Name == "mutated" {
		t.Error("repo leaked its internal slice; expected a copy")
	}
}

func TestFacilityRepoEmpty(t *testing.T) {
	repo := memory.NewFacilityRepo()
	got, err := repo.List(context.Background())
	if err != nil || len(got) != 0 {
		t.Fatalf("want empty list, got %v err %v", got, err)
	}
}
