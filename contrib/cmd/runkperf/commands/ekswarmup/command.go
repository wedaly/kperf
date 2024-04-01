package ekswarmup

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Azure/kperf/api/types"
	"github.com/Azure/kperf/contrib/internal/utils"

	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// Command represents ekswarmup subcommand.
var Command = cli.Command{
	Name:  "ekswarmup",
	Usage: "Warmup EKS cluster and try best to scale it to 8 cores at least",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "kubeconfig",
			Usage: "Path to the kubeconfig file",
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
		cli.Float64Flag{
			Name:  "rate",
			Usage: "The maximum requests per second per runner (There are 10 runners totally)",
			Value: 20,
		},
		cli.IntFlag{
			Name:  "total",
			Usage: "Total requests per runner (There are 10 runners totally and runner's rate is 20)",
			Value: 10000,
		},
	},
	Action: func(cliCtx *cli.Context) (retErr error) {
		ctx := context.Background()

		rgCfgFile, rgCfgFileDone, err := utils.NewLoadProfileFromEmbed(
			"loadprofile/ekswarmup.yaml",
			func(spec *types.RunnerGroupSpec) error {
				reqs := cliCtx.Int("total")
				if reqs < 0 {
					return fmt.Errorf("invalid total value: %v", reqs)
				}

				rate := cliCtx.Float64("rate")
				if rate <= 0 {
					return fmt.Errorf("invalid rate value: %v", rate)
				}

				spec.Profile.Spec.Total = reqs
				spec.Profile.Spec.Rate = rate

				data, _ := yaml.Marshal(spec)
				klog.V(2).InfoS("Load Profile", "config", string(data))
				return nil
			},
		)
		if err != nil {
			return err
		}
		defer func() { _ = rgCfgFileDone() }()

		kubeCfgPath := cliCtx.String("kubeconfig")

		perr := patchEKSDaemonsetWithoutToleration(ctx, kubeCfgPath)
		if perr != nil {
			return perr
		}

		cores, ferr := fetchAPIServerCores(ctx, kubeCfgPath)
		if ferr == nil {
			if isReady(cores) {
				klog.V(0).Infof("apiserver resource is ready: %v", cores)
				return nil
			}
		} else {
			klog.V(0).ErrorS(ferr, "failed to fetch apiserver cores")
		}

		delNP, err := deployWarmupVirtualNodepool(ctx, kubeCfgPath)
		if err != nil {
			return err
		}
		defer func() {
			derr := delNP()
			if retErr == nil {
				retErr = derr
			}
		}()

		var wg sync.WaitGroup
		wg.Add(1)

		jobCtx, jobCancel := context.WithCancel(ctx)
		go func() {
			defer wg.Done()

			utils.RepeatJobWith3KPod(jobCtx, kubeCfgPath, "warmupjob", 5*time.Second)
		}()

		_, derr := utils.DeployRunnerGroup(ctx,
			kubeCfgPath,
			cliCtx.String("runner-image"),
			rgCfgFile,
			cliCtx.String("runner-flowcontrol"),
			"",
		)
		jobCancel()
		wg.Wait()

		cores, ferr = fetchAPIServerCores(ctx, kubeCfgPath)
		if ferr == nil {
			if isReady(cores) {
				klog.V(0).Infof("apiserver resource is ready: %v", cores)
				return nil
			}
		}
		return derr
	},
}

// isReady returns true if there are more than two instances using 8 cores.
func isReady(cores map[string]int) bool {
	n := 0
	for _, c := range cores {
		if c >= 8 {
			n++
		}
	}
	return n >= 2
}

// deployWarmupVirtualNodepool deploys nodepool on m4.2xlarge nodes for warmup.
func deployWarmupVirtualNodepool(ctx context.Context, kubeCfgPath string) (func() error, error) {
	target := "warmup"
	kr := utils.NewKperfRunner(kubeCfgPath, "")

	klog.V(0).InfoS("Deploying virtual nodepool", "name", target)
	sharedProviderID, err := utils.FetchNodeProviderIDByType(ctx, kubeCfgPath, utils.EKSIdleNodepoolInstanceType)
	if err != nil {
		return nil, fmt.Errorf("failed to get placeholder providerID: %w", err)
	}

	klog.V(0).InfoS("Trying to delete", "nodepool", target)
	if err = kr.DeleteNodepool(ctx, 0, target); err != nil {
		klog.V(0).ErrorS(err, "failed to delete", "nodepool", target)
	}

	err = kr.NewNodepool(ctx, 0, target, 100, 110,
		"node.kubernetes.io/instance-type=m4.2xlarge", sharedProviderID)
	if err != nil {
		return nil, fmt.Errorf("failed to create nodepool %s: %w", target, err)
	}

	return func() error {
		return kr.DeleteNodepool(ctx, 0, target)
	}, nil
}

// patchEKSDaemonsetWithoutToleration removes tolerations to avoid pod scheduled
// to virtual nodes.
func patchEKSDaemonsetWithoutToleration(ctx context.Context, kubeCfgPath string) error {
	clientset := mustClientset(kubeCfgPath)

	ds := clientset.AppsV1().DaemonSets("kube-system")
	for _, dn := range []string{"aws-node", "kube-proxy"} {
		d, err := ds.Get(ctx, dn, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get daemonset %s: %w", dn, err)
		}

		d.Spec.Template.Spec.Tolerations = []corev1.Toleration{}
		d.ResourceVersion = ""

		_, err = ds.Update(ctx, d, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete toleration for daemonset %s: %w", dn, err)
		}
	}
	return nil
}

// fetchAPIServerCores fetchs core number for each kube-apiserver.
func fetchAPIServerCores(ctx context.Context, kubeCfgPath string) (map[string]int, error) {
	klog.V(0).Info("Fetching apiserver's cores")

	kr := utils.NewKubectlRunner(kubeCfgPath, "")
	fqdn, err := kr.FQDN(ctx, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster fqdn: %w", err)
	}

	ips, nerr := utils.NSLookup(fqdn)
	if nerr != nil {
		return nil, fmt.Errorf("failed get dns records of fqdn %s: %w", fqdn, nerr)
	}

	res := map[string]int{}
	for _, ip := range ips {
		cores, err := func() (int, error) {
			data, err := kr.Metrics(ctx, 0, fqdn, ip)
			if err != nil {
				return 0, fmt.Errorf("failed to get metrics for ip %s: %w", ip, err)
			}

			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "go_sched_gomaxprocs_threads") {
					vInStr := strings.Fields(line)[1]
					v, err := strconv.Atoi(vInStr)
					if err != nil {
						return 0, fmt.Errorf("failed to parse go_sched_gomaxprocs_threads %s: %w", line, err)
					}
					return v, nil
				}
			}
			return 0, fmt.Errorf("failed to get go_sched_gomaxprocs_threads")
		}()
		if err != nil {
			klog.V(0).ErrorS(err, "failed to get cores", "ip", ip)
			continue
		}
		klog.V(0).InfoS("apiserver cores", ip, cores)
		res[ip] = cores
	}
	if len(res) < 2 {
		return nil, fmt.Errorf("expected two instances at least")
	}
	return res, nil
}

// mustClientset returns kubernetes clientset without error.
func mustClientset(kubeCfgPath string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", kubeCfgPath)
	if err != nil {
		panic(fmt.Errorf("failed to build client-go config: %w", err))
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Errorf("failed to build client-go rest client: %w", err))
	}
	return clientset
}
