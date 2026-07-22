package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics - 请求总数、响应时间、错误率
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "litellm_http_requests_total",
			Help: "Total number of HTTP requests by path, method, and status code",
		},
		[]string{"path", "method", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "litellm_http_request_duration_seconds",
			Help:    "HTTP request latency distributions",
			Buckets: prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
		},
		[]string{"path", "method"},
	)

	// Provider metrics - 上游调用次数、成功/失败计数
	ProviderCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "litellm_provider_calls_total",
			Help: "Total number of provider calls by provider, model, and status",
		},
		[]string{"provider", "model", "status"},
	)

	ProviderCallDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "litellm_provider_call_duration_seconds",
			Help:    "Provider call latency distributions",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60}, // LLM 调用通常较慢
		},
		[]string{"provider", "model"},
	)

	// Token usage metrics - Token 使用统计
	TokensUsed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "litellm_tokens_used_total",
			Help: "Total tokens used by provider, model, and type (prompt/completion)",
		},
		[]string{"provider", "model", "type"},
	)
)

// RecordHTTPRequest 记录 HTTP 请求指标
func RecordHTTPRequest(path, method, status string, duration time.Duration) {
	HTTPRequestsTotal.WithLabelValues(path, method, status).Inc()
	HTTPRequestDuration.WithLabelValues(path, method).Observe(duration.Seconds())
}

// RecordProviderCall 记录 provider 调用指标
func RecordProviderCall(provider, model, status string, duration time.Duration) {
	ProviderCallsTotal.WithLabelValues(provider, model, status).Inc()
	ProviderCallDuration.WithLabelValues(provider, model).Observe(duration.Seconds())
}

// RecordTokenUsage 记录 token 使用指标
func RecordTokenUsage(provider, model string, promptTokens, completionTokens int) {
	if promptTokens > 0 {
		TokensUsed.WithLabelValues(provider, model, "prompt").Add(float64(promptTokens))
	}
	if completionTokens > 0 {
		TokensUsed.WithLabelValues(provider, model, "completion").Add(float64(completionTokens))
	}
}
