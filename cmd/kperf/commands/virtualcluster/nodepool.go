package virtualcluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/kperf/cmd/kperf/commands/utils"
	"github.com/Azure/kperf/virtualcluster"
	"helm.sh/helm/v3/pkg/release"

	"github.com/urfave/cli"
)

var nodepoolCommand = cli.Command{
	Name:  "nodepool",
	Usage: "Manage virtual node pools",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "kubeconfig",
			Usage: "Path to the kubeconfig file",
			Value: utils.DefaultKubeConfigPath,
		},
	},
	Subcommands: []cli.Command{
		nodepoolAddCommand,
		nodepoolDelCommand,
		nodepoolListCommand,
	},
}

var nodepoolAddCommand = cli.Command{
	Name:      "add",
	Usage:     "Add a virtual node pool",
	ArgsUsage: "NAME",
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "nodes",
			Usage: "The number of virtual nodes",
			Value: 10,
		},
		cli.IntFlag{
			Name:  "cpu",
			Usage: "The allocatable CPU resource per node",
			Value: 8,
		},
		cli.IntFlag{
			Name:  "memory",
			Usage: "The allocatable Memory resource per node (GiB)",
			Value: 16,
		},
		cli.StringSliceFlag{
			Name:  "affinity",
			Usage: "Deploy controllers to the nodes with a specific labels (FORMAT: KEY=VALUE[,VALUE])",
		},
	},
	Action: func(cliCtx *cli.Context) error {
		if cliCtx.NArg() != 1 {
			return fmt.Errorf("required only one argument as nodepool name")
		}
		nodepoolName := strings.TrimSpace(cliCtx.Args().Get(0))
		if len(nodepoolName) == 0 {
			return fmt.Errorf("required non-empty nodepool name")
		}

		kubeCfgPath := cliCtx.String("kubeconfig")

		affinityLabels, err := utils.KeyValuesMap(cliCtx.StringSlice("affinity"))
		if err != nil {
			return fmt.Errorf("failed to parse affinity: %w", err)
		}

		return virtualcluster.CreateNodepool(context.Background(),
			kubeCfgPath,
			nodepoolName,
			virtualcluster.WithNodepoolCPUOpt(cliCtx.Int("cpu")),
			virtualcluster.WithNodepoolMemoryOpt(cliCtx.Int("memory")),
			virtualcluster.WithNodepoolCountOpt(cliCtx.Int("nodes")),
			virtualcluster.WithNodepoolNodeControllerAffinity(affinityLabels),
		)
	},
}

var nodepoolDelCommand = cli.Command{
	Name:      "delete",
	ShortName: "del",
	ArgsUsage: "NAME",
	Usage:     "Delete a virtual node pool",
	Action: func(cliCtx *cli.Context) error {
		if cliCtx.NArg() != 1 {
			return fmt.Errorf("required only one argument as nodepool name")
		}
		nodepoolName := strings.TrimSpace(cliCtx.Args().Get(0))
		if len(nodepoolName) == 0 {
			return fmt.Errorf("required non-empty nodepool name")
		}

		kubeCfgPath := cliCtx.String("kubeconfig")

		return virtualcluster.DeleteNodepool(context.Background(), kubeCfgPath, nodepoolName)
	},
}

var nodepoolListCommand = cli.Command{
	Name:  "list",
	Usage: "List virtual node pools",
	Action: func(cliCtx *cli.Context) error {
		kubeCfgPath := cliCtx.String("kubeconfig")
		nodepools, err := virtualcluster.ListNodepools(context.Background(), kubeCfgPath)
		if err != nil {
			return err
		}
		return renderRunnerGroups(nodepools)

	},
}

func renderRunnerGroups(nodepools []*release.Release) error {
	if len(nodepools) > 0 {
		fmt.Println("+-------------------+------------+-------------+-------------+------------+")
		fmt.Printf("| %-17s | %-10s | %-9s | %-11s | %-9s |\n", "Name", "Nodes", "CPU (cores)", "Memory (GiB)", "Status")
		fmt.Println("+-------------------+------------+-------------+-------------+------------+")
	}
	for _, nodepool := range nodepools {
		fmt.Printf("| %-17s | %-10v | %-12v| %-12v| %-10v |\n", nodepool.Name, nodepool.Config["replicas"], nodepool.Config["cpu"], nodepool.Config["memory"], nodepool.Info.Status)
		fmt.Println("+-------------------+------------+-------------+-------------+------------+")
	}
	return nil
}
