# Observability

## Tracing (GEC-90)
OpenTelemetry is wired in `internal/observability`. Set the standard
`OTEL_EXPORTER_OTLP_ENDPOINT` (OTLP/HTTP) to export spans to a collector (Tempo,
Jaeger, Honeycomb, …); the whole HTTP surface is instrumented via `otelhttp`.
With no endpoint set, tracing is a zero-overhead no-op.

## Metrics (GEC-91)
The API exposes Prometheus metrics at `GET /metrics`:
`http_requests_total{route,method,status}` and
`http_request_duration_seconds_bucket{route}`. Import
`grafana-dashboard.json` into Grafana (Prometheus datasource) for request rate,
5xx error rate, and p95 latency by route.

_Follow-up:_ AI token-count/cost metrics require threading the Anthropic/Voyage
response `usage` through the outbound adapters (a metrics port); the request-level
latency above already covers Ask/Brief timing.
