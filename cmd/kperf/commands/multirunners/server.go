package multirunners

import (
	"fmt"
	"strings"

	"github.com/Azure/kperf/runner"

	"github.com/urfave/cli"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var serverCommand = cli.Command{
	Name: "server",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "namespace",
			Usage: "The namespace scope for runners",
			Value: "default",
		},
		cli.StringSliceFlag{
			Name:     "runners",
			Usage:    "The runner spec's URI",
			Required: true,
		},
		cli.StringFlag{
			Name:     "runner-image",
			Usage:    "The runner's conainer image",
			Required: true,
		},
		cli.IntFlag{
			Name:  "port",
			Value: 8080,
		},
		cli.StringFlag{
			Name:  "host",
			Value: "0.0.0.0",
		},
		cli.StringFlag{
			Name:  "data",
			Usage: "The runner result should be stored in that path",
			Value: "/tmp/data",
		},
	},
	Hidden: true,
	Action: func(cliCtx *cli.Context) error {
		name := strings.TrimSpace(cliCtx.Args().Get(0))
		if len(name) == 0 {
			return fmt.Errorf("required non-empty name")
		}

		addr := fmt.Sprintf("%s:%d", cliCtx.String("host"), cliCtx.Int("port"))
		dataDir := cliCtx.String("data")

		kubeCfgPath := cliCtx.String("kubeconfig")
		config, err := clientcmd.BuildConfigFromFlags("", kubeCfgPath)
		if err != nil {
			return err
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return err
		}

		groups := []*runner.GroupHandler{}
		imgRef := cliCtx.String("runner-image")
		ns := cliCtx.String("namespace")
		for idx, specUri := range cliCtx.StringSlice("runners") {
			gName := fmt.Sprintf("%s-%d", name, idx)
			g, err := runner.NewGroupHandler(clientset, gName, ns, specUri, imgRef)
			if err != nil {
				return err
			}
			groups = append(groups, g)
		}

		srv, err := runner.NewServer(dataDir, addr, groups...)
		if err != nil {
			return err
		}
		return srv.Run()
	},
}
