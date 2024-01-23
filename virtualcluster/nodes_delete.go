package virtualcluster

import (
	"context"
	"fmt"

	"github.com/Azure/kperf/helmcli"
)

// DeleteNodepool deletes a node pool with a given name.
func DeleteNodepool(_ context.Context, kubeconfigPath string, nodepoolName string) error {
	delCli, err := helmcli.NewDeleteCli(kubeconfigPath, virtualnodeReleaseNamespace)
	if err != nil {
		return fmt.Errorf("failed to create helm delete client: %w", err)
	}

	return delCli.Delete(nodepoolName)
}
