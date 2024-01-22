package types

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
	Total int `json:"total"`
	// List of failures
	FailureList []error `json:"failureList,omitempty"`
	// Duration means the time of benchmark.
	Duration string `json:"duration"`
	// All the observed latencies
	Latencies []float64 `json:"latencies,omitempty"`
	// total bytes read from apiserver
	TotalReceivedBytes int64 `json:"totalReceivedBytes"`
	// PercentileLatencies represents the latency distribution in seconds.
	PercentileLatencies [][2]float64 `json:"percentileLatencies,omitempty"`
}

// TODO(weifu): build brand new struct for RunnerGroupsReport to include more
// information, like how many runner groups, service account and flow control.
type RunnerGroupsReport = RunnerMetricReport
