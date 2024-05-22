package bench

import (
	kperfcmdutils "github.com/Azure/kperf/cmd/kperf/commands/utils"

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
		cli.StringFlag{
			Name:  "result",
			Usage: "Path to the file which stores results",
		},
		cli.IntFlag{
			Name:  "nodes",
			Usage: "The number of virtual nodes",
			Value: 100,
		},
		cli.IntFlag{
			Name:  "cpu",
			Usage: "The allocatable CPU resource per node",
			Value: 32,
		},
		cli.IntFlag{
			Name:  "memory",
			Usage: "The allocatable Memory resource per node (GiB)",
			Value: 96,
		},
		cli.IntFlag{
			Name:  "max-pods",
			Usage: "The maximum Pods per node",
			Value: 110,
		},
	},
	Subcommands: []cli.Command{
		benchNode100Job1Pod3KCase,
		benchNode100DeploymentNPod10KCase,
	},
}
