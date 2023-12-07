package runner

import (
	"context"
	"fmt"
	"os"
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
	},
	Action: func(cliCtx *cli.Context) error {
		profileCfg, err := loadConfig(cliCtx)
		if err != nil {
			return err
		}

		kubeCfgPath := cliCtx.String("kubeconfig")
		userAgent := cliCtx.String("user-agent")

		conns := profileCfg.Spec.Conns
		rate := profileCfg.Spec.Rate
		restClis, err := request.NewClients(kubeCfgPath, conns, userAgent, rate)
		if err != nil {
			return err
		}

		stats, err := request.Schedule(context.TODO(), &profileCfg.Spec, restClis)
		if err != nil {
			return err
		}
		printResponseStats(stats)
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
	//
	// TODO(weifu): do not override if flag is not set
	profileCfg.Spec.Rate = cliCtx.Int("rate")
	profileCfg.Spec.Conns = cliCtx.Int("conns")
	profileCfg.Spec.Total = cliCtx.Int("total")

	if err := profileCfg.Validate(); err != nil {
		return nil, err
	}
	return &profileCfg, nil
}

// printResponseStats prints ResponseStats into stdout.
func printResponseStats(stats *types.ResponseStats) {
	fmt.Println("Response stat:")
	fmt.Printf("  Total: %v\n", stats.Total)
	fmt.Printf("  Failures: %v\n", stats.Failures)
	fmt.Printf("  Duration: %v\n", stats.Duration)
	fmt.Printf("  Requests/sec: %.2f\n", float64(stats.Total)/stats.Duration.Seconds())

	fmt.Println("  Latency Distribution:")
	keys := make([]float64, 0, len(stats.Latencies))
	for q := range stats.Latencies {
		keys = append(keys, q)
	}
	sort.Float64s(keys)

	for _, q := range keys {
		fmt.Printf("    [%.2f] %.3fs\n", q, stats.Latencies[q])
	}
}
