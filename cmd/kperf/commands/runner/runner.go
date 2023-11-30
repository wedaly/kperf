package runner

import (
	"fmt"

	"github.com/urfave/cli"
)

// Command represents runner sub-command.
//
// Subcommand runner is to create request load to apiserver.
//
// NOTE: It can work with subcommand multirunners. The multirunners subcommand
// will deploy subcommand runner in pod. Details in ../multirunners.
//
// Command line interface:
//
// kperf runner --help
//
// Options:
//
//	--kubeconfig  PATH   (default: empty_string, use token if it's empty)
//	--load-config PATH   (default: empty_string, required, the config defined in api/types/load_traffic.go)
//	--conns       INT    (default: 1, Total number of connections. It can override corresponding value defined by --load-config)
//	--rate        INT    (default: 0, Maximum requests per second. It can override corresponding value defined by --load-config)
//	--total       INT    (default: 1000, Total number of request. It can override corresponding value defined by --load-config)
var Command = cli.Command{
	Name:  "runner",
	Usage: "run a load test to kube-apiserver",
	Flags: []cli.Flag{},
	Action: func(cliCtx *cli.Context) error {
		// 1. Parse options
		// 2. Setup producer-consumer goroutines
		//   2.1 Use go limter to generate request
		//   2.2 Use client-go's client to file requests
		// 3. Build progress tracker to track failure number and P99/P95/P90 latencies.
		// 4. Export summary in stdout.
		return fmt.Errorf("runner - not implemented")
	},
}
