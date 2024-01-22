package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/kperf/api/types"
	"github.com/Azure/kperf/helmcli"
	"github.com/Azure/kperf/manifests"

	"gopkg.in/yaml.v2"
)

// CreateRunnerGroupServer creates a long running server to deploy runner groups.
//
// TODO:
// 1. create a new package to define ErrNotFound, ErrAlreadyExists, ... errors.
// 2. support configurable timeout.
func CreateRunnerGroupServer(ctx context.Context, kubeconfigPath string, runnerImage string, rgSpec *types.RunnerGroupSpec) error {
	specInStr, err := tweakAndMarshalSpec(rgSpec)
	if err != nil {
		return err
	}

	getCli, err := helmcli.NewGetCli(kubeconfigPath, runnerGroupReleaseNamespace)
	if err != nil {
		return fmt.Errorf("failed to create helm get client: %w", err)
	}

	_, err = getCli.Get(runnerGroupServerReleaseName)
	if err == nil {
		return fmt.Errorf("runner group server already exists")
	}

	ch, err := manifests.LoadChart(runnerGroupServerChartName)
	if err != nil {
		return fmt.Errorf("failed to load runner group server chart: %w", err)
	}

	releaseCli, err := helmcli.NewReleaseCli(
		kubeconfigPath,
		runnerGroupReleaseNamespace,
		runnerGroupServerReleaseName,
		ch,
		runnerGroupReleaseLabels,
		helmcli.StringPathValuesApplier(
			"name="+runnerGroupServerReleaseName,
			"image="+runnerImage,
			"runnerGroupSpec="+specInStr,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create helm release client: %w", err)
	}
	return releaseCli.Deploy(ctx, 120*time.Second)
}

// tweakAndMarshalSpec updates spec's service account if not set and marshals
// it into string.
func tweakAndMarshalSpec(spec *types.RunnerGroupSpec) (string, error) {
	// NOTE: It should be aligned with ../manifests/runnergroup/server/templates/pod.yaml.
	if spec.ServiceAccount == nil {
		var sa = runnerGroupServerReleaseName
		spec.ServiceAccount = &sa
	}

	data, err := yaml.Marshal(spec)
	if err != nil {
		return "", fmt.Errorf("failed to marshal spec: %w", err)
	}
	return string(data), nil
}
