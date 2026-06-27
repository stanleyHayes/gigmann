# Service Level Objectives (GEC-94)

Targets for the Gigmann Executive Cockpit API, measured from the Prometheus
metrics at `/metrics`. Alerting rules: `alert-rules.yml`.

| SLO | Target (28-day) | Indicator |
|---|---|---|
| **Availability** | 99.0% | 1 − (5xx rate ÷ total rate) |
| **Latency (non-brief)** | p95 < 1s | `http_request_duration_seconds` p95 |
| **Brief freshness** | < TTL (10m) | brief cache refresh (background) |

Error budget: 1% of requests/28 days. The `HighErrorRate`/`HighLatencyP95`/`ApiDown`
rules fire toward these. The Daily Brief is excluded from the latency SLO because
the model call is off the hot path (served from cache).

**To activate:** point Prometheus at `/metrics`, load `alert-rules.yml`, and wire an
Alertmanager receiver (Slack/email/PagerDuty). The receiver/channel is the only
piece that needs live infrastructure + an account.
