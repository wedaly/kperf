package types

import apitypes "github.com/Azure/kperf/api/types"

// BenchmarkReport represents runkperf-bench's result.
type BenchmarkReport struct {
	apitypes.RunnerGroupsReport
	// Info is additional information.
	//
	// FIXME(weifu): Use struct after finialized.
	Info map[string]interface{}
}
