package runner

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"

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

		kubeCfgPath := cliCtx.String("kubeconfig")
		userAgent := cliCtx.String("user-agent")
		outputFile := cliCtx.String("result")

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

		var fileDir string = "result" //change directory to store response stats if needed

		var f *os.File = os.Stdout
		if outputFile != "" {
			err = os.MkdirAll(fileDir, 0750)
			if err != nil {
				log.Fatal(err)
			}
			filePath := fmt.Sprintf("%s/%s", fileDir, outputFile)
			f, err = os.Create(filePath)
			if err != nil {
				log.Fatal(err)
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
	if v := "total"; cliCtx.IsSet(v) || profileCfg.Spec.Total == 0 {
		profileCfg.Spec.Total = cliCtx.Int(v)
	}

	if err := profileCfg.Validate(); err != nil {
		return nil, err
	}
	return &profileCfg, nil
}

func printResponseStats(f *os.File, stats *types.ResponseStats) {
	fmt.Fprintf(f, "Response Stat: \n")

	fmt.Fprintf(f, "  Total: "+strconv.Itoa(stats.Total)+"\n")

	fmt.Fprintf(f, "  Total Failures: "+strconv.Itoa(len(stats.FailureList))+"\n")

	fmt.Fprintf(f, "  Observed Bytes: "+strconv.FormatInt(stats.TotalReceivedBytes, 10)+"\n")

	fmt.Fprintf(f, "  Duration: "+stats.Duration.String()+"\n")

	requestsPerSec := float64(stats.Total) / stats.Duration.Seconds()

	roundedNumber := math.Round(requestsPerSec*100) / 100
	fmt.Fprintf(f, "  Requests/sec: "+strconv.FormatFloat(roundedNumber, 'f', -1, 64)+"\n")

	fmt.Fprintf(f, "  Latency Distribution:\n")
	keys := make([]float64, 0, len(stats.PercentileLatencies))
	for q := range stats.PercentileLatencies {
		keys = append(keys, q)
	}

	sort.Float64s(keys)

	for _, q := range keys {
		str := fmt.Sprintf("    [%.2f] %.3fs\n", q/100.0, stats.PercentileLatencies[q])
		fmt.Fprint(f, str)
	}
}
