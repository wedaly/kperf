package runnergroup

import (
	"fmt"

	"github.com/urfave/cli"
)

// Command represents runnergroup sub-command.
var Command = cli.Command{
	Name:      "runnergroup",
	ShortName: "rg",
	Usage:     "deploy multiple runner groups into kubernetes",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "kubeconfig",
			Usage: "Path to the kubeconfig file",
		},
	},
	Subcommands: []cli.Command{
		runCommand,
		deleteCommand,
		waitCommand,
		resultCommand,
		serverCommand,
	},
}

var waitCommand = cli.Command{
	Name:  "wait",
	Usage: "wait until jobs finish",
	Flags: []cli.Flag{},
	Action: func(cliCtx *cli.Context) error {
		// 1. Check the progress tracker name
		// 2. Wait for the jobs
		return fmt.Errorf("wait - not implemented")
	},
}

var resultCommand = cli.Command{
	Name:  "result",
	Usage: "show the result",
	Flags: []cli.Flag{},
	Action: func(cliCtx *cli.Context) error {
		// 1. Check the progress tracker name
		// 2. Ensure the jobs finished
		// 3. Output the result
		return fmt.Errorf("result - not implemented")
	},
}
