package main

import (
	"fmt"
	"os"

	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/projecteru2/cli/commands"
	"github.com/projecteru2/cli/utils"
	"github.com/projecteru2/cli/versioninfo"
	"gopkg.in/urfave/cli.v2"
)

var (
	debug   bool
	eru     string
	timeout int
)

func setupLog(l string) error {
	level, err := log.ParseLevel(l)
	if err != nil {
		return err
	}
	log.SetLevel(level)

	formatter := &log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}
	log.SetFormatter(formatter)
	return nil
}

func subCommand(c *cli.Context, n string) bool {
	return strings.HasSuffix(c.App.Name, n)
}

func run(c *cli.Context) error {
	if debug {
		setupLog("DEBUG")
	} else {
		setupLog("INFO")
	}
	conn := utils.ConnectEru(eru, timeout)
	if subCommand(c, "raw") && c.Command.Name == "deploy" {
		commands.RawDeploy(c, conn)
	} else {
		log.Fatal("Not support yet")
	}
	return nil
}

func init() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Print(versioninfo.VersionString())
	}
}

func main() {
	app := cli.App{}
	app.Name = versioninfo.NAME
	app.Usage = "control eru in shell"
	app.Version = versioninfo.VERSION

	rawCommand := &cli.Command{
		Name:  "raw",
		Usage: "use it to deploy/rm containers by raw image",
		Subcommands: []*cli.Command{
			{
				Name:      "deploy",
				Usage:     "deploy raw image by specs",
				ArgsUsage: "a spec URL",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "pod",
						Usage: "where to run",
					},
					&cli.StringFlag{
						Name:  "entry",
						Usage: "which entry",
					},
					&cli.StringFlag{
						Name:  "image",
						Usage: "which to run",
					},
					&cli.IntFlag{
						Name:  "count",
						Usage: "how many",
						Value: 1,
					},
					&cli.StringFlag{
						Name:  "network",
						Usage: "SDN name or host mode",
						Value: "host",
					},
					&cli.Float64Flag{
						Name:  "cpu",
						Usage: "how many cpu",
						Value: 1.0,
					},
					&cli.Int64Flag{
						Name:  "mem",
						Usage: "how many memory in bytes",
						Value: 536870912.0,
					},
					&cli.StringSliceFlag{
						Name:  "env",
						Usage: "set env can use multiple times",
					},
				},
				Action: run,
			},
		},
	}
	app.Commands = []*cli.Command{
		rawCommand,
	}
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "debug",
			Usage:       "enable debug",
			Aliases:     []string{"d"},
			Value:       false,
			Destination: &debug,
		},
		&cli.StringFlag{
			Name:        "eru",
			Usage:       "eru core address",
			Aliases:     []string{"e"},
			Value:       "127.0.0.1:5001",
			EnvVars:     []string{"ERU"},
			Destination: &eru,
		},
		&cli.IntFlag{
			Name:        "timeout",
			Usage:       "timeout for conn",
			Aliases:     []string{"t"},
			Value:       2,
			Destination: &timeout,
		},
	}
	app.Run(os.Args)
}
