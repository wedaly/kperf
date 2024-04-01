package utils

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/Azure/kperf/api/types"
	"github.com/Azure/kperf/contrib/internal/manifests"

	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	// EKSIdleNodepoolInstanceType is the instance type of idle nodepool.
	//
	// NOTE: The EKS cloud provider will delete all the NOT-READY nodes
	// which aren't managed by it. When kwok-controller fails to update
	// virtual node's lease, the EKS cloud provider would delete that
	// virtual node. It's unexpected behavior. In order to avoid this case,
	// we should create a idle nodepool with one node and use that node's
	// provider ID for all the virtual nodes so that EKS cloud provider
	// won't delete our virtual nodes.
	EKSIdleNodepoolInstanceType = "m4.large"

	// EKSRunnerNodepoolInstanceType is the instance type of nodes for kperf
	// runners.
	//
	// NOTE: This is default type. Please align it with ../manifests/loadprofile/ekswarmup.yaml.
	EKSRunnerNodepoolInstanceType = "m4.4xlarge"
)

// RepeatJobWith3KPod repeats to deploy 3k pods.
func RepeatJobWith3KPod(ctx context.Context, kubeCfgPath string, namespace string, internal time.Duration) {
	klog.V(0).Info("Repeat to create job with 3k pods")

	target := "workload/3kpod.job.yaml"

	data, err := manifests.FS.ReadFile(target)
	if err != nil {
		panic(fmt.Errorf("unexpected error when read %s from embed memory: %v",
			target, err))
	}

	jobFile, cleanup, err := CreateTempFileWithContent(data)
	if err != nil {
		panic(fmt.Errorf("unexpected error when create job yaml: %v", err))
	}
	defer func() { _ = cleanup() }()

	kr := NewKubectlRunner(kubeCfgPath, namespace)

	klog.V(0).Infof("Creating namespace %s", namespace)
	err = kr.CreateNamespace(ctx, 5*time.Minute, namespace)
	if err != nil {
		panic(fmt.Errorf("failed to create a new namespace %s: %v", namespace, err))
	}

	defer func() {
		klog.V(0).Infof("Cleanup namespace %s", namespace)
		err := kr.DeleteNamespace(context.TODO(), 5*time.Minute, namespace)
		if err != nil {
			klog.V(0).ErrorS(err, "failed to cleanup", "namespace", namespace)
		}
	}()

	retryInterval := 5 * time.Second
	for {
		select {
		case <-ctx.Done():
			klog.V(0).Info("Stop creating job")
			return
		default:
		}

		time.Sleep(retryInterval)

		aerr := kr.Apply(ctx, 5*time.Minute, jobFile)
		if aerr != nil {
			klog.V(0).ErrorS(aerr, "failed to apply, retry after 5 seconds", "job", target)
			continue
		}

		werr := kr.Wait(ctx, 15*time.Minute, "condition=complete", "15m", "job/batchjobs")
		if werr != nil {
			klog.V(0).ErrorS(werr, "failed to wait", "job", target)
		}

		derr := kr.Delete(ctx, 5*time.Minute, jobFile)
		if derr != nil {
			klog.V(0).ErrorS(derr, "failed to delete", "job", target)
		}
		time.Sleep(internal)
	}
}

// NewLoadProfileFromEmbed reads load profile from embed memory.
func NewLoadProfileFromEmbed(target string, tweakFn func(*types.RunnerGroupSpec) error) (_name string, _cleanup func() error, _ error) {
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

	return CreateTempFileWithContent(data)
}

// DeployRunnerGroup deploys runner group for benchmark.
func DeployRunnerGroup(ctx context.Context,
	kubeCfgPath, runnerImage, rgCfgFile string,
	runnerFlowControl, runnerGroupAffinity string) (string, error) {

	klog.V(0).InfoS("Deploying runner group", "config", rgCfgFile)

	kr := NewKperfRunner(kubeCfgPath, runnerImage)

	klog.V(0).Info("Deleting existing runner group")
	derr := kr.RGDelete(ctx, 0)
	if derr != nil {
		klog.V(0).ErrorS(derr, "failed to delete existing runner group")
	}

	rerr := kr.RGRun(ctx, 0, rgCfgFile, runnerFlowControl, runnerGroupAffinity)
	if rerr != nil {
		return "", fmt.Errorf("failed to deploy runner group: %w", rerr)
	}

	klog.V(0).Info("Waiting runner group")
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
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
		return res, nil
	}
}

// FetchNodeProviderIDByType is used to get one node's provider id with a given
// instance type.
func FetchNodeProviderIDByType(ctx context.Context, kubeCfgPath string, instanceType string) (string, error) {
	clientset, err := buildClientset(kubeCfgPath)
	if err != nil {
		return "", err
	}

	label := fmt.Sprintf("node.kubernetes.io/instance-type=%v", instanceType)

	nodeCli := clientset.CoreV1().Nodes()
	listResp, err := nodeCli.List(ctx, metav1.ListOptions{LabelSelector: label})
	if err != nil {
		return "", fmt.Errorf("failed to list nodes with label %s: %w", label, err)
	}

	if len(listResp.Items) == 0 {
		return "", fmt.Errorf("there is no such node with label %s", label)
	}
	return listResp.Items[0].Spec.ProviderID, nil
}

// buildClientset returns kubernetes clientset.
func buildClientset(kubeCfgPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeCfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build client-go config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build client-go rest client: %w", err)
	}
	return clientset, nil
}

// NSLookup returns ips for URL.
func NSLookup(domainURL string) ([]string, error) {
	ips, err := net.LookupHost(domainURL)
	if err != nil {
		return nil, err
	}
	sort.Strings(ips)
	return ips, nil
}

// runCommand runs command with Pdeathsig.
func runCommand(ctx context.Context, timeout time.Duration, cmd string, args []string) ([]byte, error) {
	var cancel context.CancelFunc
	if timeout != 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	c := exec.CommandContext(ctx, cmd, args...)
	c.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}

	klog.V(2).Infof("[CMD] %s", c.String())
	output, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to invoke %s:\n (output: %s): %w",
			c.String(), strings.TrimSpace(string(output)), err)
	}
	return output, nil
}

// CreateTempFileWithContent creates temporary file with data.
func CreateTempFileWithContent(data []byte) (_name string, _cleanup func() error, retErr error) {
	f, err := os.CreateTemp("", "temp*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	fName := f.Name()
	defer func() {
		if retErr != nil {
			_ = os.RemoveAll(fName)
		}
	}()

	_, err = f.Write(data)
	f.Close()
	if err != nil {
		return "", nil, fmt.Errorf("failed to write temporary in %s: %w",
			fName, err)
	}

	return fName, func() error {
		return os.RemoveAll(fName)
	}, nil
}
