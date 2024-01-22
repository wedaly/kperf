package multirunners

import (
	"context"
	"fmt"

	"github.com/Azure/kperf/api/types"
	"github.com/Azure/kperf/runner"
	runnergroup "github.com/Azure/kperf/runner/group"

	"github.com/urfave/cli"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: "run runner groups",
	Flags: []cli.Flag{
		// TODO(weifu): need https://github.com/Azure/kperf/issues/25 to support list
		cli.StringSliceFlag{
			Name:     "runnergroup",
			Usage:    "The runner group spec's URI",
			Required: true,
		},
		cli.StringFlag{
			Name:  "runner-image",
			Usage: "The runner's conainer image",
			// TODO(weifu):
			//
			// We should build release pipeline so that we can
			// build with fixed public release image as default value.
			// Right now, we need to set image manually.
			Required: true,
		},
	},
	Action: func(cliCtx *cli.Context) error {
		imgRef := cliCtx.String("runner-image")
		if len(imgRef) == 0 {
			return fmt.Errorf("required valid runner image")
		}

		specs, err := loadRunnerGroupSpec(cliCtx)
		if err != nil {
			return fmt.Errorf("failed to load runner group spec: %w", err)
		}
		if len(specs) != 1 {
			return fmt.Errorf("only support one runner group right now. will support it after https://github.com/Azure/kperf/issues/25")
		}

		kubeCfgPath := cliCtx.GlobalString("kubeconfig")
		return runner.CreateRunnerGroupServer(context.Background(),
			kubeCfgPath,
			imgRef,
			specs[0],
		)
	},
}

// loadRunnerGroupSpec loads runner group spec from URIs.
func loadRunnerGroupSpec(cliCtx *cli.Context) ([]*types.RunnerGroupSpec, error) {
	clientset, err := buildKubernetesClientset(cliCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubernetes clientset: %w", err)
	}

	specURIs := cliCtx.StringSlice("runnergroup")

	specs := make([]*types.RunnerGroupSpec, 0, len(specURIs))
	for _, specURI := range specURIs {
		spec, err := runnergroup.NewRunnerGroupSpecFromURI(clientset, specURI)
		if err != nil {
			return nil, err
		}

		specs = append(specs, spec)
	}
	return specs, nil
}
