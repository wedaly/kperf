package runnergroup

import (
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
		resultCommand,
		serverCommand,
		statusCommand,
	},
}
