package virtualcluster

import (
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/release"

	"github.com/Azure/kperf/helmcli"
)

// ListNodeppol lists nodepools added by the vc nodeppool add command.
func ListNodepools(_ context.Context, kubeconfigPath string) ([]*release.Release, error) {
	listCli, err := helmcli.NewListCli(kubeconfigPath, virtualnodeReleaseNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create helm list client: %w", err)
	}

	return listCli.List()

}
