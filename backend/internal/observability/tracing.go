// Package observability wires OpenTelemetry tracing. It is infrastructure (used by
// the composition root), never imported by core/app/ports.
package observability

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// SetupTracing configures the global OpenTelemetry TracerProvider to export spans
// over OTLP/HTTP. The endpoint is read from the standard
// OTEL_EXPORTER_OTLP_ENDPOINT / OTEL_EXPORTER_OTLP_TRACES_ENDPOINT env vars; when
// neither is set, tracing is a no-op (the default global provider) so the demo
// runs with zero tracing overhead. Returns a shutdown func to flush on exit.
func SetupTracing(ctx context.Context, serviceName, env string) (func(context.Context) error, error) {
	noop := func(context.Context) error { return nil }
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" && os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") == "" {
		return noop, nil
	}

	exp, err := otlptracehttp.New(ctx) // endpoint + headers come from OTEL_* env
	if err != nil {
		return nil, fmt.Errorf("observability: otlp trace exporter: %w", err)
	}
	res, err := resource.Merge(resource.Default(), resource.NewSchemaless(
		attribute.String("service.name", serviceName),
		attribute.String("deployment.environment", env),
	))
	if err != nil {
		return nil, fmt.Errorf("observability: resource: %w", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))
	return tp.Shutdown, nil
}
