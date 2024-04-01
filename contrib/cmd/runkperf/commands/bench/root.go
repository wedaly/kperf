package bench

import (
	"context"
	"fmt"

	kperfcmdutils "github.com/Azure/kperf/cmd/kperf/commands/utils"
	"github.com/Azure/kperf/contrib/internal/utils"
	"k8s.io/klog/v2"

	"github.com/urfave/cli"
)

// Command represents bench subcommand.
var Command = cli.Command{
	Name:  "bench",
	Usage: "Run benchmark test cases",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "kubeconfig",
			Usage: "Path to the kubeconfig file",
			Value: kperfcmdutils.DefaultKubeConfigPath,
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
		cli.StringFlag{
			Name:  "runner-flowcontrol",
			Usage: "Apply flowcontrol to runner group. (FORMAT: PriorityLevel:MatchingPrecedence)",
			Value: "workload-low:1000",
		},
		cli.StringFlag{
			Name:  "vc-affinity",
			Usage: "Deploy virtualnode's controller with a specific labels (FORMAT: KEY=VALUE[,VALUE])",
			Value: "node.kubernetes.io/instance-type=Standard_D8s_v3,m4.2xlarge",
		},
		cli.StringFlag{
			Name:  "rg-affinity",
			Usage: "Deploy runner group with a specific labels (FORMAT: KEY=VALUE[,VALUE])",
			Value: "node.kubernetes.io/instance-type=Standard_D16s_v3,m4.4xlarge",
		},
		cli.BoolFlag{
			Name:   "eks",
			Usage:  "Indicates the target kubernetes cluster is EKS",
			Hidden: true,
		},
	},
	Subcommands: []cli.Command{
		benchNode100Job1Pod3KCase,
	},
}

// deployVirtualNodepool deploys virtual nodepool.
func deployVirtualNodepool(ctx context.Context, cliCtx *cli.Context, target string, nodes, maxPods int) (func() error, error) {
	klog.V(0).InfoS("Deploying virtual nodepool", "name", target)

	kubeCfgPath := cliCtx.GlobalString("kubeconfig")
	virtualNodeAffinity := cliCtx.GlobalString("vc-affinity")

	kr := utils.NewKperfRunner(kubeCfgPath, "")

	var sharedProviderID string
	var err error

	if cliCtx.GlobalBool("eks") {
		sharedProviderID, err = utils.FetchNodeProviderIDByType(ctx, kubeCfgPath, utils.EKSIdleNodepoolInstanceType)
		if err != nil {
			return nil, fmt.Errorf("failed to get EKS idle node (type: %s) providerID: %w",
				utils.EKSIdleNodepoolInstanceType, err)
		}
	}

	klog.V(0).InfoS("Trying to delete nodepool if necessary", "name", target)
	if err = kr.DeleteNodepool(ctx, 0, target); err != nil {
		klog.V(0).ErrorS(err, "failed to delete nodepool", "name", target)
	}

	err = kr.NewNodepool(ctx, 0, target, nodes, maxPods, virtualNodeAffinity, sharedProviderID)
	if err != nil {
		return nil, fmt.Errorf("failed to create nodepool %s: %w", target, err)
	}

	return func() error {
		return kr.DeleteNodepool(ctx, 0, target)
	}, nil
}
