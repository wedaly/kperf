package bench

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/kperf/api/types"
	kperfcmdutils "github.com/Azure/kperf/cmd/kperf/commands/utils"
	"github.com/Azure/kperf/contrib/internal/manifests"
	"github.com/Azure/kperf/contrib/internal/utils"
	"k8s.io/klog/v2"

	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
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

// deployRunnerGroup deploys runner group for benchmark.
func deployRunnerGroup(ctx context.Context, cliCtx *cli.Context, rgCfgFile string) error {
	klog.V(0).InfoS("Deploying runner group", "config", rgCfgFile)

	kubeCfgPath := cliCtx.GlobalString("kubeconfig")
	runnerImage := cliCtx.GlobalString("runner-image")

	kr := utils.NewKperfRunner(kubeCfgPath, runnerImage)

	klog.V(0).Info("Deleting existing runner group")
	derr := kr.RGDelete(ctx, 0)
	if derr != nil {
		klog.V(0).ErrorS(derr, "failed to delete existing runner group")
	}

	runnerFlowControl := cliCtx.GlobalString("runner-flowcontrol")
	runnerGroupAffinity := cliCtx.GlobalString("rg-affinity")

	rerr := kr.RGRun(ctx, 0, rgCfgFile, runnerFlowControl, runnerGroupAffinity)
	if rerr != nil {
		return fmt.Errorf("failed to deploy runner group: %w", rerr)
	}

	klog.V(0).Info("Waiting runner group")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		res, err := kr.RGResult(ctx, 1*time.Minute)
		if err != nil {
			klog.V(0).ErrorS(err, "failed to fetch warmup runner group's result")
			continue
		}
		klog.V(0).InfoS("Runner group's result", "data", res)

		klog.V(0).Info("Deleting runner group")
		if derr := kr.RGDelete(ctx, 0); derr != nil {
			klog.V(0).ErrorS(err, "failed to delete runner group")
		}
		return nil
	}
}

// newLoadProfileFromEmbed reads load profile from embed memory.
func newLoadProfileFromEmbed(target string, tweakFn func(*types.RunnerGroupSpec) error) (_name string, _cleanup func() error, _ error) {
	data, err := manifests.FS.ReadFile(target)
	if err != nil {
		return "", nil, fmt.Errorf("unexpected error when read %s from embed memory: %v", target, err)
	}

	if tweakFn != nil {
		var spec types.RunnerGroupSpec
		if err = yaml.Unmarshal(data, &spec); err != nil {
			return "", nil, fmt.Errorf("failed to unmarshal into runner group spec:\n (data: %s)\n: %w",
				string(data), err)
		}

		if err = tweakFn(&spec); err != nil {
			return "", nil, err
		}

		data, err = yaml.Marshal(spec)
		if err != nil {
			return "", nil, fmt.Errorf("failed to marshal runner group spec after tweak: %w", err)
		}
	}

	return utils.CreateTempFileWithContent(data)
}
