package metrics

import (
	"container/list"
	"sync"
	"sync/atomic"

	"github.com/Azure/kperf/api/types"
)

// ResponseMetric is a measurement related to http response.
type ResponseMetric interface {
	// ObserveLatency observes latency.
	ObserveLatency(seconds float64)
	// ObserveFailure observes failure response.
	ObserveFailure(err error)
	// ObserveReceivedBytes observes the bytes read from apiserver.
	ObserveReceivedBytes(bytes int64)
	// Gather returns the summary.
	Gather() types.ResponseStats
}

type responseMetricImpl struct {
	mu            sync.Mutex
	failureList   []error
	latencies     *list.List
	receivedBytes int64
}

func NewResponseMetric() ResponseMetric {
	errList := make([]error, 0, 1024)
	return &responseMetricImpl{
		latencies:   list.New(),
		failureList: errList,
	}
}

// ObserveLatency implements ResponseMetric.
func (m *responseMetricImpl) ObserveLatency(seconds float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.latencies.PushBack(seconds)
}

// ObserveFailure implements ResponseMetric.
func (m *responseMetricImpl) ObserveFailure(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failureList = append(m.failureList, err)
}

// ObserveReceivedBytes implements ResponseMetric.
func (m *responseMetricImpl) ObserveReceivedBytes(bytes int64) {
	atomic.AddInt64(&m.receivedBytes, bytes)
}

// Gather implements ResponseMetric.
func (m *responseMetricImpl) Gather() types.ResponseStats {
	latencies := m.dumpLatencies()
	return types.ResponseStats{
		FailureList:        m.failureList,
		Latencies:          latencies,
		TotalReceivedBytes: atomic.LoadInt64(&m.receivedBytes),
	}
}

func (m *responseMetricImpl) dumpLatencies() []float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make([]float64, 0, m.latencies.Len())
	for e := m.latencies.Front(); e != nil; e = e.Next() {
		res = append(res, e.Value.(float64))
	}
	return res
}
