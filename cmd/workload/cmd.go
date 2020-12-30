package workload

import (
	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/core/strategy"
	"github.com/urfave/cli/v2"
)

const (
	workloadArgsUsage = "workloadID(s)"
	specFileURI       = "<spec file uri>"
	copyArgsUsage     = "workloadID:path1,path2,...,pathn"
	sendArgsUsage     = "path1,path2,...pathn"
)

// Command exports workload subommands
func Command() *cli.Command {
	return &cli.Command{
		Name:    "workload",
		Aliases: []string{"container"},
		Usage:   "workload commands",
		Subcommands: []*cli.Command{
			{
				Name:      "get",
				Usage:     "get workload(s)",
				ArgsUsage: workloadArgsUsage,
				Action:    utils.ExitCoder(cmdWorkloadGet),
			},
			{
				Name:      "logs",
				Usage:     "get workload stream logs",
				ArgsUsage: "workloadID",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "tail",
						Value: "all",
						Usage: `number of lines to show from the end of the logs (default "all")`,
					},
					&cli.StringFlag{
						Name:  "since",
						Usage: "show logs since timestamp (e.g. 2013-01-02T13:23:37) or relative (e.g. 42m for 42 minutes)",
					},
					&cli.StringFlag{
						Name:  "until",
						Usage: "show logs before a timestamp (e.g. 2013-01-02T13:23:37) or relative (e.g. 42m for 42 minutes)",
					},
					&cli.BoolFlag{
						Name:    "follow",
						Aliases: []string{"f"},
						Usage:   "follow log output",
					},
				},
				Action: utils.ExitCoder(cmdWorkloadLogs),
			},
			{
				Name:      "get-status",
				Usage:     "get workload status",
				ArgsUsage: workloadArgsUsage,
				Action:    utils.ExitCoder(cmdWorkloadGetStatus),
			},
			{
				Name:      "set-status",
				Usage:     "set workload status",
				ArgsUsage: workloadArgsUsage,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "running",
						Usage: "Running",
					},
					&cli.BoolFlag{
						Name:  "healthy",
						Usage: "Healthy",
					},
					&cli.Int64Flag{
						Name:  "ttl",
						Usage: "ttl",
						Value: 0,
					},
					&cli.StringSliceFlag{
						Name:  "network",
						Usage: "network, can set multiple times, name=ip",
					},
					&cli.StringFlag{
						Name:  "extension",
						Usage: "extension things",
					},
				},
				Action: utils.ExitCoder(cmdWorkloadSetStatus),
			},
			{
				Name:      "list",
				Usage:     "list workload(s) by appname",
				ArgsUsage: "[appname]",
				Action:    utils.ExitCoder(cmdWorkloadList),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "entry",
						Usage: "filter by entry",
					},
					&cli.StringFlag{
						Name:  "nodename",
						Usage: "filter by nodename",
					},
					&cli.StringSliceFlag{
						Name:  "label",
						Usage: "label filter can set multiple times",
					},
					&cli.Int64Flag{
						Name:  "limit",
						Usage: "limit data size",
					},
				},
			},
			{
				Name:      "stop",
				Usage:     "stop workload(s)",
				ArgsUsage: workloadArgsUsage,
				Action:    utils.ExitCoder(cmdWorkloadStop),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Usage:   "force to stop",
						Aliases: []string{"f"},
						Value:   false,
					},
				},
			},
			{
				Name:      "start",
				Usage:     "start workload(s)",
				ArgsUsage: workloadArgsUsage,
				Action:    utils.ExitCoder(cmdWorkloadStart),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Usage:   "force to start",
						Aliases: []string{"f"},
						Value:   false,
					},
				},
			},
			{
				Name:      "restart",
				Usage:     "restart workload(s)",
				ArgsUsage: workloadArgsUsage,
				Action:    utils.ExitCoder(cmdWorkloadRestart),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Usage:   "force to restart",
						Aliases: []string{"f"},
						Value:   false,
					},
				},
			},
			{
				Name:      "remove",
				Usage:     "remove workload(s)",
				ArgsUsage: workloadArgsUsage,
				Action:    utils.ExitCoder(cmdWorkloadRemove),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Usage:   "force to remove",
						Aliases: []string{"f"},
						Value:   false,
					},
					&cli.IntFlag{
						Name:    "step",
						Usage:   "concurrent remove step",
						Aliases: []string{"s"},
						Value:   1,
					},
				},
			},
			{
				Name:      "copy",
				Usage:     "copy file(s) from workload(s)",
				ArgsUsage: copyArgsUsage,
				Action:    utils.ExitCoder(cmdWorkloadCopy),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "dir",
						Usage:   "where to store",
						Aliases: []string{"d"},
						Value:   "/tmp",
					},
				},
			},
			{
				Name:      "send",
				Usage:     "send file(s) to workload(s)",
				ArgsUsage: sendArgsUsage,
				Action:    utils.ExitCoder(cmdWorkloadSend),
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:  "file",
						Usage: "copy local files to workload, can use multiple times. src_path:dst_path",
					},
				},
			},
			{
				Name:      "dissociate",
				Usage:     "dissociate workload(s) from eru, return it resource but not remove it",
				ArgsUsage: workloadArgsUsage,
				Action:    utils.ExitCoder(cmdWorkloadDissociate),
			},
			{
				Name:      "realloc",
				Usage:     "realloc workloads resource",
				ArgsUsage: workloadArgsUsage,
				Action:    utils.ExitCoder(cmdWorkloadRealloc),
				Flags: []cli.Flag{
					&cli.Float64Flag{
						Name:  "cpu-request",
						Usage: "cpu request increment/decrement",
						Value: 0,
					},
					&cli.Float64Flag{
						Name:  "cpu-limit",
						Usage: "cpu limit increment/decrement",
						Value: 0,
					},
					&cli.StringFlag{
						Name:  "memory-request",
						Usage: "memory request increment/decrement, like -1M or 1G, support K, M, G, T",
					},
					&cli.StringFlag{
						Name:  "memory-limit",
						Usage: "memory limit increment/decrement, like -1M or 1G, support K, M, G, T",
					},
					&cli.StringFlag{
						Name:  "volumes-request",
						Usage: `volumes request increment/decrement, like "AUTO:/data:rw:-1G,/tmp:/tmp"`,
					},
					&cli.StringFlag{
						Name:  "volumes-limit",
						Usage: `volumes limit increment/decrement, like "AUTO:/data:rw:-1G,/tmp:/tmp"`,
					},
					&cli.BoolFlag{
						Name:  "cpu-bind",
						Usage: `bind fixed cpu(s) with workload`,
					},
					&cli.BoolFlag{
						Name:  "cpu-unbind",
						Usage: `unbind the workload relation with cpu`,
					},
					&cli.StringFlag{
						Name:  "storage-request",
						Usage: `storage request incr/decr, like "-1G"`,
					},
					&cli.StringFlag{
						Name:  "storage-limit",
						Usage: `storage limit incr/decr, like "-1G"`,
					},
				},
			},
			{
				Name:      "exec",
				Usage:     "run a command in a running workload",
				ArgsUsage: "workloadID -- cmd1 cmd2 cmd3",
				Action:    utils.ExitCoder(cmdWorkloadExec),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "interactive",
						Aliases: []string{"i"},
						Value:   false,
					},
					&cli.StringSliceFlag{
						Name:    "env",
						Aliases: []string{"e"},
						Usage:   "ENV=value",
					},
					&cli.StringFlag{
						Name:    "workdir",
						Aliases: []string{"w"},
						Usage:   "/path/to/workdir",
						Value:   "/",
					},
				},
			},
			{
				Name:      "replace",
				Usage:     "replace workloads by params",
				ArgsUsage: specFileURI,
				Action:    utils.ExitCoder(cmdWorkloadReplace),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "pod",
						Usage: "where to replace",
					},
					&cli.StringFlag{
						Name:  "entry",
						Usage: "which entry",
					},
					&cli.StringFlag{
						Name:  "image",
						Usage: "which to replace",
					},
					&cli.StringFlag{
						Name:  "node",
						Usage: "which node to replace",
						Value: "",
					},
					&cli.IntFlag{
						Name:  "count",
						Usage: "run simultaneously",
						Value: 1,
					},
					&cli.BoolFlag{
						Name:  "network-inherit",
						Usage: "use old workload network configuration",
						Value: false,
					},
					&cli.StringFlag{
						Name:  "network",
						Usage: "SDN name or host mode",
						//	Value: "host",
					},
					&cli.StringSliceFlag{
						Name:  "env",
						Usage: "set env can use multiple times, e.g., GO111MODULE=on",
					},
					&cli.StringFlag{
						Name:  "user",
						Usage: "which user",
						Value: "root",
					},
					&cli.StringSliceFlag{
						Name:  "label",
						Usage: "filter workload by labels",
					},
					&cli.StringSliceFlag{
						Name:  "file",
						Usage: "copy local files to workload, can use multiple times. src_path:dst_path",
					},
					&cli.StringSliceFlag{
						Name:  "copy",
						Usage: "copy old workload files to new workload, can use multiple times. src_path:dst_path",
					},
					&cli.BoolFlag{
						Name:  "debug",
						Usage: "enable debug mode for workload send their logs to default log driver",
					},
					&cli.BoolFlag{
						Name:  "ignore-hook",
						Usage: "ignore-hook result",
						Value: false,
					},
					&cli.StringSliceFlag{
						Name:  "after-create",
						Usage: "run commands after create",
					},
				},
			},
			{
				Name:      "deploy",
				Usage:     "deploy workloads by params",
				ArgsUsage: specFileURI,
				Action:    utils.ExitCoder(cmdWorkloadDeploy),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "dry run show capacity",
					},
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
					&cli.StringSliceFlag{
						Name:  "node",
						Usage: "which node to run",
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
						Name:  "cpu-request",
						Usage: "how many cpu to request",
						Value: 0,
					},
					&cli.Float64Flag{
						Name:  "cpu-limit",
						Usage: "how many cpu to limit; can specify limit without request",
						Value: 1.0,
					},
					&cli.StringFlag{
						Name:  "memory-request",
						Usage: "how many memory to request like 1M or 1G, support K, M, G, T",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "memory-limit",
						Usage: "how many memory to limit like 1M or 1G, support K, M, G, T; can specify limit without request",
						Value: "512M",
					},
					&cli.StringFlag{
						Name:  "storage-request",
						Usage: "how many storage to request quota like 1M or 1G, support K, M, G, T",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "storage-limit",
						Usage: "how many storage to limit quota like 1M or 1G, support K, M, G, T; can specify limit without request",
						Value: "",
					},
					&cli.StringSliceFlag{
						Name:  "env",
						Usage: "set env can use multiple times, e.g., GO111MODULE=on",
					},
					&cli.StringSliceFlag{
						Name:  "nodelabel",
						Usage: "filter nodes by labels",
					},
					&cli.StringFlag{
						Name:  "deploy-strategy",
						Usage: "deploy method auto/fill/each/global/fillglobal",
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
					&cli.StringSliceFlag{
						Name:  "after-create",
						Usage: "run commands after create",
					},
					&cli.BoolFlag{
						Name:  "debug",
						Usage: "enable debug mode for workload send their logs to default log driver",
					},
					&cli.IntFlag{
						Name:  "nodes-limit",
						Usage: "Limit nodes count in fill and each mode",
						Value: 0,
					},
					&cli.BoolFlag{
						Name:  "auto-replace",
						Usage: "create or replace automatically",
					},
					&cli.BoolFlag{
						Name:  "cpu-bind",
						Usage: "bind cpu or not",
						Value: false,
					},
					&cli.BoolFlag{
						Name:  "ignore-hook",
						Usage: "ignore hook process",
						Value: false,
					},
					&cli.StringFlag{
						Name:  "raw-args",
						Usage: "raw args in json (for docker engine)",
						Value: "",
					},
				},
			},
		},
	}
}
