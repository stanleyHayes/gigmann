package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/localnarrator"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/signal"
)

func TestStaticBriefGenerate(t *testing.T) {
	asOf := time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC)
	svc := app.NewBriefService(signal.Default(signal.DefaultThresholds()), localnarrator.New(), 5)
	gen := app.NewStaticBrief(svc, briefInput(asOf))

	b, err := gen.Generate(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, b.Items, "the synthetic network should yield brief items")
	assert.Contains(t, b.Prose, "Sammy")
}
