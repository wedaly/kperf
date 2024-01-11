package types

import "time"

// ResponseStats is the report about benchmark result.
type ResponseStats struct {
	// List of failures
	FailureList []error
	// All the observed latencies
	Latencies []float64
	// total bytes read from apiserver
	TotalReceivedBytes int64
}

type RunnerMetricReport struct {
	// Total represents total number of requests.
	Total int
	// List of failures
	FailureList []error
	// Duration means the time of benchmark.
	Duration time.Duration
	// All the observed latencies
	Latencies []float64
	// total bytes read from apiserver
	TotalReceivedBytes int64
	// PercentileLatencies represents the latency distribution in seconds.
	PercentileLatencies [][2]float64 // [2]float64{percentile, value}
}
