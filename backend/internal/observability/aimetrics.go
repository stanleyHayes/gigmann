package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// aiReg is a dedicated registry for AI (Claude/Voyage) usage metrics, exposed
// alongside the HTTP metrics at /metrics.
var aiReg = prometheus.NewRegistry()

var (
	aiRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "ai_requests_total", Help: "AI generation calls by operation and outcome."},
		[]string{"op", "outcome"},
	)
	aiTokens = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "ai_tokens_total", Help: "AI tokens consumed by operation and kind (input/output)."},
		[]string{"op", "kind"},
	)
	aiDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "ai_request_duration_seconds", Help: "AI generation latency by operation.", Buckets: prometheus.DefBuckets},
		[]string{"op"},
	)
)

func init() { aiReg.MustRegister(aiRequests, aiTokens, aiDuration) }

// AIRegistry returns the registry holding AI usage metrics (for the /metrics handler).
func AIRegistry() *prometheus.Registry { return aiReg }

// RecordAICall records one AI generation call: token counts, latency, and outcome.
// Token counts ≤ 0 are skipped (e.g. the deterministic local fallback reports none).
func RecordAICall(op string, inputTokens, outputTokens int, dur time.Duration, err error) {
	outcome := "success"
	if err != nil {
		outcome = "error"
	}
	aiRequests.WithLabelValues(op, outcome).Inc()
	aiDuration.WithLabelValues(op).Observe(dur.Seconds())
	if inputTokens > 0 {
		aiTokens.WithLabelValues(op, "input").Add(float64(inputTokens))
	}
	if outputTokens > 0 {
		aiTokens.WithLabelValues(op, "output").Add(float64(outputTokens))
	}
}
