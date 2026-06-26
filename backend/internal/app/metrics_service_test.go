package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/memory"
	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/seed"
)

func TestMetricsServiceNetwork(t *testing.T) {
	net := seed.Generate(7, time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC), 14)
	svc := app.NewMetricsService(memory.NewMetricsRepo(net.Metrics...))

	n, err := svc.Network(context.Background())
	require.NoError(t, err)
	require.Len(t, n.KPIs, 4)
	assert.Equal(t, "revenue", n.KPIs[0].Key)
}
