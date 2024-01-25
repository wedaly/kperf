package runner

import (
	"context"
	"encoding/json"

	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/kperf/api/types"
	"github.com/Azure/kperf/metrics"
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
		cli.Float64Flag{
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
		cli.BoolFlag{
			Name:  "raw-data",
			Usage: "show raw letencies data in result",
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
		rawDataFlagIncluded := cliCtx.Bool("raw-data")

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
			outputFileDir := filepath.Dir(outputFilePath)

			_, err = os.Stat(outputFileDir)
			if err != nil && os.IsNotExist(err) {
				err = os.MkdirAll(outputFileDir, 0750)
			}
			if err != nil {
				return fmt.Errorf("failed to ensure output's dir %s: %w", outputFileDir, err)
			}

			f, err = os.Create(outputFilePath)
			if err != nil {
				return err
			}
			defer f.Close()
		}

		err = printResponseStats(f, rawDataFlagIncluded, stats)
		if err != nil {
			return fmt.Errorf("error while printing response stats: %w", err)
		}

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
		profileCfg.Spec.Rate = cliCtx.Float64(v)
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

// printResponseStats prints types.RunnerMetricReport into underlying file.
func printResponseStats(f *os.File, rawDataFlagIncluded bool, stats *request.Result) error {
	output := types.RunnerMetricReport{
		Total:               stats.Total,
		FailureList:         stats.FailureList,
		Duration:            stats.Duration.String(),
		Latencies:           stats.Latencies,
		TotalReceivedBytes:  stats.TotalReceivedBytes,
		PercentileLatencies: metrics.BuildPercentileLatencies(stats.Latencies),
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")

	if !rawDataFlagIncluded {
		output.Latencies = nil
	}

	err := encoder.Encode(output)
	if err != nil {
		return fmt.Errorf("failed to encode json: %w", err)
	}
	return nil
}
