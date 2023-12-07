package metrics

import (
	"fmt"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
)

// ResponseMetric is a measurement related to http response.
type ResponseMetric interface {
	// ObserveLatency observes latency.
	ObserveLatency(seconds float64)
	// ObserveFailure observes failure response.
	ObserveFailure()
	// Gather returns the summary.
	Gather() (latencies map[float64]float64, failure int, _ error)
}

type responseMetricImpl struct {
	latencySeconds *prometheus.SummaryVec
	failureCount   int64
}

func NewResponseMetric() ResponseMetric {
	return &responseMetricImpl{
		latencySeconds: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:  "request",
				Name:       "request_latency_seconds",
				Objectives: map[float64]float64{0: 0, 0.5: 0, 0.9: 0, 0.95: 0, 0.99: 0, 1: 0},
			},
			[]string{},
		),
	}
}

// ObserveLatency implements ResponseMetric.
func (m *responseMetricImpl) ObserveLatency(seconds float64) {
	m.latencySeconds.WithLabelValues().Observe(seconds)
}

// ObserveFailure implements ResponseMetric.
func (m *responseMetricImpl) ObserveFailure() {
	atomic.AddInt64(&m.failureCount, 1)
}

// Gather implements ResponseMetric.
func (m *responseMetricImpl) Gather() (map[float64]float64, int, error) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(m.latencySeconds)

	metricFamilies, err := reg.Gather()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to gather from local registry: %w", err)
	}

	latencies := map[float64]float64{}
	for _, q := range metricFamilies[0].GetMetric()[0].GetSummary().GetQuantile() {
		latencies[q.GetQuantile()] = q.GetValue()
	}

	return latencies, int(atomic.LoadInt64(&m.failureCount)), nil
}
