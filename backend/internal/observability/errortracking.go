package observability

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
)

// SetupErrorTracking initialises Sentry error reporting. With an empty DSN the SDK
// is disabled (events are dropped), so this is a no-op in the demo/offline path.
// Returns a flush func to call on shutdown.
func SetupErrorTracking(dsn, env string) (func(), error) {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:           dsn, // empty → SDK disabled
		Environment:   env,
		EnableTracing: false, // tracing is handled by OpenTelemetry
	}); err != nil {
		return nil, fmt.Errorf("observability: sentry init: %w", err)
	}
	return func() { sentry.Flush(2 * time.Second) }, nil
}
