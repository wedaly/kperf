package bench

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Azure/kperf/api/types"
	kperfcmdutils "github.com/Azure/kperf/cmd/kperf/commands/utils"
	"github.com/Azure/kperf/contrib/internal/utils"

	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

var benchNode100Job1Pod3KCase = cli.Command{
	Name: "node100_job1_pod3k",
	Usage: `

The test suite is to setup 100 virtual nodes and deploy one job with 3k pods on
that nodes. It repeats to create and delete job. The load profile is fixed.
	`,
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "total-requests",
			Usage: "Total requests per runner (There are 10 runners totally and runner's rate is 10)",
			Value: 36000,
		},
	},
	Action: func(cliCtx *cli.Context) error {
		ctx := context.Background()
		kubeCfgPath := cliCtx.GlobalString("kubeconfig")

		rgCfgFile, rgCfgFileDone, err := utils.NewLoadProfileFromEmbed(
			"loadprofile/node100_job1_pod3k.yaml",
			func(spec *types.RunnerGroupSpec) error {
				reqs := cliCtx.Int("total-requests")
				if reqs < 0 {
					return fmt.Errorf("invalid total-requests value: %v", reqs)
				}

				rgAffinity := cliCtx.GlobalString("rg-affinity")
				affinityLabels, err := kperfcmdutils.KeyValuesMap([]string{rgAffinity})
				if err != nil {
					return fmt.Errorf("failed to parse %s affinity: %w", rgAffinity, err)
				}

				spec.Profile.Spec.Total = reqs
				spec.NodeAffinity = affinityLabels

				data, _ := yaml.Marshal(spec)
				klog.V(2).InfoS("Load Profile", "config", string(data))
				return nil
			},
		)
		if err != nil {
			return err
		}
		defer func() { _ = rgCfgFileDone() }()

		vcDone, err := deployVirtualNodepool(ctx, cliCtx, "node100job1pod3k", 100, 110)
		if err != nil {
			return fmt.Errorf("failed to deploy virtual node: %w", err)
		}
		defer func() { _ = vcDone() }()

		var wg sync.WaitGroup
		wg.Add(1)

		jobCtx, jobCancel := context.WithCancel(ctx)
		go func() {
			defer wg.Done()

			utils.RepeatJobWith3KPod(jobCtx, kubeCfgPath, "job1pod3k", 5*time.Second)
		}()

		// TODO(weifu): write result into file
		_, derr := utils.DeployRunnerGroup(ctx,
			cliCtx.GlobalString("kubeconfig"),
			cliCtx.GlobalString("runner-image"),
			rgCfgFile,
			cliCtx.GlobalString("runner-flowcontrol"),
			cliCtx.GlobalString("rg-affinity"),
		)
		jobCancel()
		wg.Wait()
		return derr
	},
}
