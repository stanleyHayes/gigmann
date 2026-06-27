package observability_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/observability"
)

func TestSetupTracingNoopWithoutEndpoint(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
	shutdown, err := observability.SetupTracing(context.Background(), "gigmann-api", "test")
	require.NoError(t, err)
	require.NotNil(t, shutdown)
	require.NoError(t, shutdown(context.Background()))
}
