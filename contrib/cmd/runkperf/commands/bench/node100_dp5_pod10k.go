package bench

import (
	"context"
	"fmt"
	"sync"
	"time"

	internaltypes "github.com/Azure/kperf/contrib/internal/types"
	"github.com/Azure/kperf/contrib/internal/utils"

	"github.com/urfave/cli"
)

var benchNode100Deployment5Pod10KCase = cli.Command{
	Name: "node100_dp5_pod10k",
	Usage: `

The test suite is to setup 100 virtual nodes and deploy 5 deployments for 10k
pods on that nodes. It repeats to rolling-update deployments one by one during
benchmark.
	`,
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "total",
			Usage: "Total requests per runner (There are 10 runners totally and runner's rate is 10)",
			Value: 36000,
		},
		cli.IntFlag{
			Name:  "podsize",
			Usage: "Add <key=data, value=randomStringByLen(podsize)> in pod's annotation to increase pod size. The value is close to pod's size",
			Value: 0,
		},
	},
	Action: func(cliCtx *cli.Context) error {
		_, err := renderBenchmarkReportInterceptor(
			addAPIServerCoresInfoInterceptor(benchNode100Deployment5Pod10KRun),
		)(cliCtx)
		return err
	},
}

// benchNode100Deployment5Pod10KCase is for subcommand benchNode100Deployment5Pod10KCase.
func benchNode100Deployment5Pod10KRun(cliCtx *cli.Context) (*internaltypes.BenchmarkReport, error) {
	ctx := context.Background()
	kubeCfgPath := cliCtx.GlobalString("kubeconfig")

	rgCfgFile, rgSpec, rgCfgFileDone, err := newLoadProfileFromEmbed(cliCtx,
		"loadprofile/node100_dp5_pod10k.yaml")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rgCfgFileDone() }()

	vcDone, err := deployVirtualNodepool(ctx, cliCtx, "node100dp5pod10k", 100, 150)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy virtual node: %w", err)
	}
	defer func() { _ = vcDone() }()

	var wg sync.WaitGroup
	wg.Add(1)

	restartInterval := 10 * time.Second
	dpCtx, dpCancel := context.WithCancel(ctx)

	podSize := cliCtx.Int("podsize")
	rollingUpdateFn, err := utils.RepeatRollingUpdate10KPod(dpCtx, kubeCfgPath, "dp5pod10k", podSize, restartInterval)
	if err != nil {
		dpCancel()
		return nil, fmt.Errorf("failed to setup workload: %w", err)
	}

	go func() {
		defer wg.Done()

		// FIXME(weifu):
		//
		// DeployRunnerGroup should return ready notification.
		// The rolling update should run after runners.
		rollingUpdateFn()
	}()

	rgResult, derr := utils.DeployRunnerGroup(ctx,
		cliCtx.GlobalString("kubeconfig"),
		cliCtx.GlobalString("runner-image"),
		rgCfgFile,
		cliCtx.GlobalString("runner-flowcontrol"),
		cliCtx.GlobalString("rg-affinity"),
	)
	dpCancel()
	wg.Wait()

	if derr != nil {
		return nil, derr
	}

	return &internaltypes.BenchmarkReport{
		Description: fmt.Sprintf(`
Environment: 100 virtual nodes managed by kwok-controller,
Workload: Deploy 5 deployments with 10,000 pods. Rolling-update deployments one by one and the interval is %v`, restartInterval),
		LoadSpec: *rgSpec,
		Result:   *rgResult,
		Info: map[string]interface{}{
			"podSizeInBytes": podSize,
		},
	}, nil
}
