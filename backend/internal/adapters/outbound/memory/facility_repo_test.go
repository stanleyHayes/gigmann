package memory_test

import (
	"context"
	"testing"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/core/facility"
)

func mustFacility(t *testing.T) facility.Facility {
	t.Helper()
	f, err := facility.New("f1", "Kasoa", "Central", "Kasoa", 40, facility.StatusGood)
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
