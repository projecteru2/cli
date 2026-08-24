package main

import (
	"context"
	"fmt"
	"os"

	"github.com/projecteru2/core/log"
	coretypes "github.com/projecteru2/core/types"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/core"
	"github.com/projecteru2/cli/cmd/image"
	"github.com/projecteru2/cli/cmd/lambda"
	"github.com/projecteru2/cli/cmd/network"
	"github.com/projecteru2/cli/cmd/node"
	"github.com/projecteru2/cli/cmd/pod"
	"github.com/projecteru2/cli/cmd/status"
	"github.com/projecteru2/cli/cmd/workload"
	"github.com/projecteru2/cli/describe"
	"github.com/projecteru2/cli/version"
)

func main() {
	cli.VersionPrinter = func(_ *cli.Command) {
		fmt.Print(version.String())
	}

	ctx := context.Background()
	if err := newApp().Run(ctx, os.Args); err != nil {
		log.WithFunc("main").Fatalf(ctx, err, "run eru-cli")
	}
}

func newApp() *cli.Command {
	app := &cli.Command{
		Name:    version.NAME,
		Usage:   "control eru in shell",
		Version: version.VERSION,
		Before:  setupLog,
		Commands: []*cli.Command{
			core.Command(),
			image.Command(),
			lambda.Command(),
			network.Command(),
			node.Command(),
			pod.Command(),
			status.Command(),
			workload.Command(),
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "enable debug",
				Aliases: []string{"d"},
				Value:   false,
			},
			&cli.StringFlag{
				Name:    "eru",
				Usage:   "eru core address",
				Aliases: []string{"e"},
				Value:   "127.0.0.1:5001",
				Sources: cli.EnvVars("ERU"),
			},
			&cli.StringFlag{
				Name:    "username",
				Usage:   "eru core username",
				Aliases: []string{"u"},
				Value:   "",
				Sources: cli.EnvVars("ERU_USERNAME"),
			},
			&cli.StringFlag{
				Name:    "password",
				Usage:   "eru core password",
				Aliases: []string{"p"},
				Value:   "",
				Sources: cli.EnvVars("ERU_PASSWORD"),
			},
			&cli.StringFlag{
				Name:        "output",
				Usage:       "output format, json / yaml",
				Aliases:     []string{"o"},
				Value:       "",
				Sources:     cli.EnvVars("ERU_OUTPUT_FORMAT"),
				Destination: &describe.Format,
			},
		},
	}

	// v3 reads the separator setting off the command owning the flag, not off the root.
	_ = app.Walk(func(cmd *cli.Command) error {
		cmd.DisableSliceFlagSeparator = true
		return nil
	})
	return app
}

func setupLog(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	level := "info"
	if cmd.Bool("debug") {
		level = "debug"
	}
	return ctx, log.SetupLog(ctx, &coretypes.ServerLogConfig{Level: level}, "")
}
