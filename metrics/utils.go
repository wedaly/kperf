package metrics

import (
	"math"
	"sort"
)

// BuildPercentileLatencies builds percentile latencies.
func BuildPercentileLatencies(latencies []float64) [][2]float64 {
	if len(latencies) == 0 {
		return nil
	}

	var percentiles = []float64{0, 0.5, 0.90, 0.95, 0.99, 1}

	res := make([][2]float64, len(percentiles))

	n := len(latencies)
	sort.Float64s(latencies)
	for pi, pv := range percentiles {
		idx := int(math.Ceil(float64(n) * pv))
		if idx > 0 {
			idx--
		}
		res[pi] = [2]float64{pv, latencies[idx]}
	}
	return res
}
