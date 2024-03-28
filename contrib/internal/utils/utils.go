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

	"github.com/Azure/kperf/contrib/internal/manifests"

	"k8s.io/klog/v2"
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
	defer cleanup()

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
			os.RemoveAll(fName)
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
