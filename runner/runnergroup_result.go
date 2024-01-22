package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Azure/kperf/api/types"
	"github.com/Azure/kperf/portforward"
)

// GetRunnerGroupResult gets runner group's aggregated report.
func GetRunnerGroupResult(_ context.Context, kubecfgPath string) (*types.RunnerGroupsReport, error) {
	pf, err := portforward.NewPodPortForwarder(
		kubecfgPath,
		runnerGroupReleaseNamespace,
		runnerGroupServerReleaseName,
		runnerGroupServerPort,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init pod portforward: %w", err)
	}
	defer pf.Stop()

	if err = pf.Start(); err != nil {
		return nil, fmt.Errorf("failed to start pod port forward: %w", err)
	}

	localPort, err := pf.GetLocalPort()
	if err != nil {
		return nil, fmt.Errorf("failed to get local port: %w", err)
	}

	targetURL := fmt.Sprintf("http://localhost:%d/v1/runnergroups/summary", localPort)

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
