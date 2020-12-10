package lambda

import (
	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/core/strategy"
	"github.com/urfave/cli/v2"
)

// Command exports lambda subommands
func Command() *cli.Command {
	return &cli.Command{
		Name:  "lambda",
		Usage: "run commands in a workload like local",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "name for this lambda",
			},
			&cli.StringFlag{
				Name:  "network",
				Usage: "SDN name",
			},
			&cli.StringFlag{
				Name:  "pod",
				Usage: "where to run",
			},
			&cli.StringSliceFlag{
				Name:  "env",
				Usage: "set env can use multiple times, e.g., GO111MODULE=on",
			},
			&cli.StringSliceFlag{
				Name:  "volume-request",
				Usage: "set volume request can use multiple times",
			},
			&cli.StringSliceFlag{
				Name:  "volume",
				Usage: "set volume limitcan use multiple times",
			},
			&cli.StringFlag{
				Name:  "working_dir",
				Usage: "use as current working dir",
				Value: "/",
			},
			&cli.StringFlag{
				Name:  "image",
				Usage: "base image for running",
				Value: "alpine:latest",
			},
			&cli.Float64Flag{
				Name:  "cpu-request",
				Usage: "how many cpu request",
				Value: 0,
			},
			&cli.Float64Flag{
				Name:  "cpu",
				Usage: "how many cpu limit",
				Value: 1.0,
			},
			&cli.StringFlag{
				Name:  "memory-request",
				Usage: "memory request, support K, M, G, T",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "memory",
				Usage: "memory limit, support K, M, G, T",
				Value: "512M",
			},
			&cli.StringFlag{
				Name:  "storage-request",
				Usage: "how many storage to request quota like 1M or 1G, support K, M, G, T",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "storage",
				Usage: "how many storage to limit quota like 1M or 1G, support K, M, G, T",
				Value: "",
			},
			&cli.IntFlag{
				Name:  "count",
				Usage: "how many workloads",
				Value: 1,
			},
			&cli.BoolFlag{
				Name:    "stdin",
				Usage:   "open stdin for workload",
				Aliases: []string{"s"},
				Value:   false,
			},
			&cli.StringFlag{
				Name:  "deploy-strategy",
				Usage: "deploy method auto/fill/each",
				Value: strategy.Auto,
			},
			&cli.StringFlag{
				Name:  "user",
				Usage: "which user",
				Value: "root",
			},
			&cli.StringSliceFlag{
				Name:  "file",
				Usage: "copy local file to workload, can use multiple times. src_path:dst_path",
			},
			&cli.BoolFlag{
				Name:  "async",
				Usage: "run lambda async",
			},
			&cli.IntFlag{
				Name:  "async-timeout",
				Usage: "for async timeout",
				Value: 30,
			},
			&cli.BoolFlag{
				Name:    "privileged",
				Usage:   "give extended privileges to this lambda",
				Aliases: []string{"p"},
				Value:   false,
			},
			&cli.StringSliceFlag{
				Name:  "node",
				Usage: "which node to run",
			},
		},
		Action: utils.ExitCoder(cmdLambdaRun),
	}
}
