package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Azure/kperf/api/types"
)

// GetRunnerGroupResult gets runner group's aggregated report.
func GetRunnerGroupResult(_ context.Context, kubecfgPath string) (*types.RunnerGroupsReport, error) {
	host, done, err := initPortForwardToServer(kubecfgPath)
	if err != nil {
		return nil, err
	}
	defer done()

	targetURL := fmt.Sprintf("http://%s/v1/runnergroups/summary", host)

	// FIXME(weifu): cleanup nolint
	//nolint:gosec
	resp, err := http.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to access %s by portforward: %w", targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errInRaw, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error message when http code = %v: %w",
				resp.Status, err)
		}

		herr := types.HTTPError{}
		err = json.Unmarshal(errInRaw, &herr)
		if err != nil {
			return nil, fmt.Errorf("failed to get error when http code = %v: %w",
				resp.Status, err)
		}
		return nil, herr
	}

	dataInRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	res := types.RunnerGroupsReport{}
	err = json.Unmarshal(dataInRaw, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal to get result: %w\n\n%s",
			err, string(dataInRaw))
	}
	return &res, nil
}
