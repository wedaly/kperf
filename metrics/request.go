package metrics

import (
	"container/list"
	"errors"
	"io"
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
	errorStats    *types.ResponseErrorStats
	latencies     *list.List
	receivedBytes int64
}

func NewResponseMetric() ResponseMetric {
	return &responseMetricImpl{
		latencies:  list.New(),
		errorStats: types.NewResponseErrorStats(),
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
	if err == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// HTTP2 -> TCP/TLS -> Unknown
	code := codeFromHTTP(err)
	switch {
	case code != 0:
		m.errorStats.ResponseCodes[code]++
	case isHTTP2Error(err):
		updateHTTP2ErrorStats(m.errorStats, err)
	case isDialTimeoutError(err) || errors.Is(err, io.ErrUnexpectedEOF) || isConnectionRefused(err):
		updateNetErrors(m.errorStats, err)
	default:
		m.errorStats.UnknownErrors = append(m.errorStats.UnknownErrors, err.Error())
	}
}

// ObserveReceivedBytes implements ResponseMetric.
func (m *responseMetricImpl) ObserveReceivedBytes(bytes int64) {
	atomic.AddInt64(&m.receivedBytes, bytes)
}

// Gather implements ResponseMetric.
func (m *responseMetricImpl) Gather() types.ResponseStats {
	return types.ResponseStats{
		ErrorStats:         m.dumpErrorStats(),
		Latencies:          m.dumpLatencies(),
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

func (m *responseMetricImpl) dumpErrorStats() types.ResponseErrorStats {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.errorStats.Copy()
}
