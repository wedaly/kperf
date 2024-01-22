package runnergroup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Azure/kperf/api/types"
	"github.com/Azure/kperf/runner"

	"github.com/urfave/cli"
)

var resultCommand = cli.Command{
	Name:  "result",
	Usage: "show the runner groups' result",
	Flags: []cli.Flag{},
	Action: func(cliCtx *cli.Context) error {
		kubeCfgPath := cliCtx.GlobalString("kubeconfig")

		res, err := runner.GetRunnerGroupResult(context.Background(), kubeCfgPath)
		if err != nil {
			return err
		}

		return renderRunnerGroupsReport(res)
	},
}

// renderRunnerGroupsReport renders runner groups' report into stdio.
func renderRunnerGroupsReport(res *types.RunnerGroupsReport) error {
	encoder := json.NewEncoder(os.Stdout)

	encoder.SetIndent("", "  ")
	err := encoder.Encode(res)
	if err != nil {
		return fmt.Errorf("failed to encode json: %w", err)
	}
	return nil
}
