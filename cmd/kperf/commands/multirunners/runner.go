package multirunners

import (
	"fmt"

	"github.com/urfave/cli"
)

// Command represents multirunners sub-command.
//
// Subcommand multirunners is to deploy multiple runners as kubernetes jobs.
// Since one runner could run out of networking bandwidth on one host, the
// multirunners deploys runners on different hosts and reduces the impact of
// limited networking resource.
//
// Command line interface:
//
// kperf mrunners run --help
//
// Options:
//
//	--kubeconfig   PATH     (default: empty_string, use token if it's empty)
//	--namespace    STRING   (default: empty_string, required)
//	--runner-image STRING   (default: empty_string, required)
//	--runners      []STRING (default: empty, required)
//	--wait         BOOLEAN  (default: false)
//
// Details:
//
// The --runners format is defined by URI.
//
// - file:///abs_path?numbers=10
// - configmap:///namespace/name?numbers=2
// - ...
//
// The schema:://PATH is used to get runner's configuration. It can be local
// path or stored as configmap in target kubernetes cluster. The query part is
// to define what the job looks like. Currently, that command just requires
// the number of pods in that job. At the beginning, we just need to file://.
// The number of runners defines the number of jobs.
//
// All the jobs are referenced by one configmap (ownerReference). The configmap
// name will be output to stdout. The name is progress tracker ID. By default,
// there is only one progress tracker ID in one namespace.
//
// kperf mrunners wait --help
//
// Args:
//
//	0: namespace (STRING)
//
// Options:
//
//	--kubeconfig   PATH     (default: empty_string, use token if it's empty)
//
// Wait it to wait until jobs finish.
//
// kperf mrunners result --help
//
// Args:
//
//	0: namespace (STRING)
//
// Options:
//
//	--kubeconfig   PATH     (default: empty_string, use token if it's empty)
//
// Result retrieves the result for jobs. If jobs is still running, that command
// will fail.
var Command = cli.Command{
	Name:      "multirunners",
	ShortName: "mrunners",
	Usage:     "packages runner as job and deploy runners into kubernetes",
	Subcommands: []cli.Command{
		runCommand,
		waitCommand,
		resultCommand,
	},
}

var runCommand = cli.Command{
	Name:  "run",
	Flags: []cli.Flag{},
	Action: func(cliCtx *cli.Context) error {
		// 1. Parse options
		// 2. Deploy jobs for --runners
		// 3. Wait
		return fmt.Errorf("run - not implemented")
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
