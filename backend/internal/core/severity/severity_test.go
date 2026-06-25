package severity_test

import (
	"testing"

	"github.com/xcreativs/gigmann/internal/core/severity"
)

func TestValid(t *testing.T) {
	for _, s := range []severity.Severity{severity.Good, severity.Watch, severity.Critical} {
		if !s.Valid() {
			t.Errorf("%q should be valid", s)
		}
	}
	if severity.Severity("nope").Valid() {
		t.Error("unknown severity reported valid")
	}
}

func TestRankOrdersWorstHighest(t *testing.T) {
	if severity.Critical.Rank() <= severity.Watch.Rank() || severity.Watch.Rank() <= severity.Good.Rank() {
		t.Errorf("rank order wrong: good=%d watch=%d critical=%d",
			severity.Good.Rank(), severity.Watch.Rank(), severity.Critical.Rank())
	}
	if severity.Severity("nope").Rank() != 0 {
		t.Error("unknown severity should rank 0")
	}
}
