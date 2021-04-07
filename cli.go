package main

import (
	"fmt"
	"os"

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
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var debug bool

func setupLog(l string) error {
	logrus.SetOutput(os.Stderr)
	level, err := logrus.ParseLevel(l)
	if err != nil {
		return err
	}
	logrus.SetLevel(level)

	formatter := &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}
	logrus.SetFormatter(formatter)
	return nil
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Print(version.String())
	}

	app := &cli.App{
		Name:    version.NAME,
		Usage:   "control eru in shell",
		Version: version.VERSION,
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
				Name:        "debug",
				Usage:       "enable debug",
				Aliases:     []string{"d"},
				Value:       false,
				Destination: &debug,
			},
			&cli.StringFlag{
				Name:    "eru",
				Usage:   "eru core address",
				Aliases: []string{"e"},
				Value:   "127.0.0.1:5001",
				EnvVars: []string{"ERU"},
			},
			&cli.StringFlag{
				Name:    "username",
				Usage:   "eru core username",
				Aliases: []string{"u"},
				Value:   "",
				EnvVars: []string{"ERU_USERNAME"},
			},
			&cli.StringFlag{
				Name:    "password",
				Usage:   "eru core password",
				Aliases: []string{"p"},
				Value:   "",
				EnvVars: []string{"ERU_PASSWORD"},
			},
			&cli.StringFlag{
				Name:        "output",
				Usage:       "output format, json / yaml",
				Aliases:     []string{"o"},
				Value:       "",
				EnvVars:     []string{"ERU_OUTPUT_FORMAT"},
				Destination: &describe.Format,
			},
		},
	}

	var loglevel string
	if debug {
		loglevel = "DEBUG"
	} else {
		loglevel = "INFO"
	}

	if err := setupLog(loglevel); err != nil {
		fmt.Printf("Error setup log: %v\n", err)
		os.Exit(-1)
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("Error running eru-cli: %v\n", err)
		os.Exit(-1)
	}
}
