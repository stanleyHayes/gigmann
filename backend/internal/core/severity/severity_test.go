package severity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xcreativs/gigmann/internal/core/severity"
)

func TestValid(t *testing.T) {
	for _, s := range []severity.Severity{severity.Good, severity.Watch, severity.Critical} {
		assert.Truef(t, s.Valid(), "%q should be valid", s)
	}
	assert.False(t, severity.Severity("nope").Valid())
}

func TestRankOrdersWorstHighest(t *testing.T) {
	assert.Greater(t, severity.Critical.Rank(), severity.Watch.Rank())
	assert.Greater(t, severity.Watch.Rank(), severity.Good.Rank())
	assert.Equal(t, 0, severity.Severity("nope").Rank())
}
