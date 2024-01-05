package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Azure/kperf/api/types"
	"github.com/Azure/kperf/request"

	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

// Command represents runner subcommand.
var Command = cli.Command{
	Name:  "runner",
	Usage: "Setup benchmark to kube-apiserver from one endpoint",
	Subcommands: []cli.Command{
		runCommand,
	},
}

var runCommand = cli.Command{
	Name:  "run",
	Usage: "run a benchmark test to kube-apiserver",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "kubeconfig",
			Usage: "Path to the kubeconfig file",
		},
		cli.IntFlag{
			Name:  "client",
			Usage: "Total number of HTTP clients",
			Value: 1,
		},
		cli.StringFlag{
			Name:     "config",
			Usage:    "Path to the configuration file",
			Required: true,
		},
		cli.IntFlag{
			Name:  "conns",
			Usage: "Total number of connections. It can override corresponding value defined by --config",
			Value: 1,
		},
		cli.StringFlag{
			Name:  "content-type",
			Usage: "Content type (json or protobuf)",
			Value: "json",
		},
		cli.IntFlag{
			Name:  "rate",
			Usage: "Maximum requests per second (Zero means no limitation). It can override corresponding value defined by --config",
		},
		cli.IntFlag{
			Name:  "total",
			Usage: "Total number of requests. It can override corresponding value defined by --config",
			Value: 1000,
		},
		cli.StringFlag{
			Name:  "user-agent",
			Usage: "User Agent",
		},
		cli.StringFlag{
			Name:  "result",
			Usage: "Path to the file which stores results",
		},
	},
	Action: func(cliCtx *cli.Context) error {
		profileCfg, err := loadConfig(cliCtx)
		if err != nil {
			return err
		}

		// Get the content type from the command-line flag
		contentType := cliCtx.String("content-type")
		kubeCfgPath := cliCtx.String("kubeconfig")
		userAgent := cliCtx.String("user-agent")
		outputFilePath := cliCtx.String("result")

		conns := profileCfg.Spec.Conns
		rate := profileCfg.Spec.Rate
		restClis, err := request.NewClients(kubeCfgPath, conns, userAgent, rate, contentType)
		if err != nil {
			return err
		}
		stats, err := request.Schedule(context.TODO(), &profileCfg.Spec, restClis)

		if err != nil {
			return err
		}

		var f *os.File = os.Stdout
		if outputFilePath != "" {
			err := os.MkdirAll(filepath.Dir(outputFilePath), 0600)
			if err != nil {
				return err
			}
			f, err = os.Create(outputFilePath)
			if err != nil {
				return err
			}
			defer f.Close()
		}
		printResponseStats(f, stats)
		return nil
	},
}

// loadConfig loads and validates the config.
func loadConfig(cliCtx *cli.Context) (*types.LoadProfile, error) {
	var profileCfg types.LoadProfile

	cfgPath := cliCtx.String("config")

	cfgInRaw, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", cfgPath, err)
	}

	if err := yaml.Unmarshal(cfgInRaw, &profileCfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s from yaml format: %w", cfgPath, err)
	}

	// override value by flags
	if v := "rate"; cliCtx.IsSet(v) {
		profileCfg.Spec.Rate = cliCtx.Int(v)
	}
	if v := "conns"; cliCtx.IsSet(v) || profileCfg.Spec.Conns == 0 {
		profileCfg.Spec.Conns = cliCtx.Int(v)
	}
	if v := "client"; cliCtx.IsSet(v) || profileCfg.Spec.Client == 0 {
		profileCfg.Spec.Client = cliCtx.Int(v)
	}
	if v := "total"; cliCtx.IsSet(v) || profileCfg.Spec.Total == 0 {
		profileCfg.Spec.Total = cliCtx.Int(v)
	}

	if err := profileCfg.Validate(); err != nil {
		return nil, err
	}
	return &profileCfg, nil
}

func printResponseStats(f *os.File, stats *types.ResponseStats) {
	fmt.Fprint(f, "Response Stat: \n")
	fmt.Fprintf(f, "  Total: %v\n", stats.Total)

	fmt.Fprintf(f, "  Total Failures: %d\n", len(stats.FailureList))

	for _, v := range stats.FailureList {
		fmt.Fprintf(f, "	%v\n", v)
	}

	fmt.Fprintf(f, "  Observed Bytes: %v\n", stats.TotalReceivedBytes)

	fmt.Fprintf(f, "  Duration: %v\n", stats.Duration.String())

	requestsPerSec := float64(stats.Total) / stats.Duration.Seconds()

	fmt.Fprintf(f, "  Requests/sec: %.2f\n", requestsPerSec)

	fmt.Fprint(f, "  Latency Distribution:\n")
	keys := make([]float64, 0, len(stats.PercentileLatencies))
	for q := range stats.PercentileLatencies {
		keys = append(keys, q)
	}

	sort.Float64s(keys)

	for _, q := range keys {
		fmt.Fprintf(f, "    [%.2f] %.3fs\n", q/100.0, stats.PercentileLatencies[q])
	}
}
