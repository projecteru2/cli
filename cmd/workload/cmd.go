package workload

import (
	corecluster "github.com/projecteru2/core/cluster"
	"github.com/projecteru2/core/strategy"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

const (
	workloadArgsUsage = "workloadID(s)"
	specFileURI       = "<spec file uri>"
	copyArgsUsage     = "workloadID:path1,path2,...,pathn"

	flagEntry   = "entry"
	flagNode    = "node"
	flagPod     = "pod"
	flagForce   = "force"
	flagFile    = "file"
	flagEnv     = "env"
	flagImage   = "image"
	flagNetwork = "network"
	flagStorage = "storage"

	flagCPURequest     = "cpu-request"
	flagCPULimit       = "cpu-limit"
	flagMemoryRequest  = "memory-request"
	flagMemoryLimit    = "memory-limit"
	flagStorageRequest = "storage-request"
	flagStorageLimit   = "storage-limit"
	flagVolumesRequest = "volumes-request"
	flagVolumesLimit   = "volumes-limit"
)

var stopOnFirstArg = 1

// Command returns the workload command tree.
func Command() *cli.Command {
	return &cli.Command{
		Name:    "workload",
		Aliases: []string{"container"},
		Usage:   "workload commands",
		Commands: []*cli.Command{
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
						Name:  flagNetwork,
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
						Name:  flagEntry,
						Usage: "filter by entry",
					},
					&cli.StringFlag{
						Name:  flagNode,
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
					&cli.StringSliceFlag{
						Name:  "match-ip",
						Usage: "filter by IP",
					},
					&cli.StringSliceFlag{
						Name:  "skip-ip",
						Usage: "filter out IP",
					},
					&cli.StringSliceFlag{
						Name:  flagPod,
						Usage: "filter by Pod",
					},
					&cli.BoolFlag{
						Name:  "statistics",
						Usage: "Display the statistics of Workloads",
					},
				},
			},
			{
				Name:      "stop",
				Usage:     "stop workload(s)",
				ArgsUsage: workloadArgsUsage,
				Action:    utils.ExitCoder(cmdWorkloadControl(corecluster.WorkloadStop)),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    flagForce,
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
				Action:    utils.ExitCoder(cmdWorkloadControl(corecluster.WorkloadStart)),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    flagForce,
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
				Action:    utils.ExitCoder(cmdWorkloadControl(corecluster.WorkloadRestart)),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    flagForce,
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
						Name:    flagForce,
						Usage:   "force to remove",
						Aliases: []string{"f"},
						Value:   false,
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
				ArgsUsage: workloadArgsUsage,
				Action:    utils.ExitCoder(cmdWorkloadSend),
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:  flagFile,
						Usage: "copy local files to workload, can use multiple times. src_path:dst_path[:mode[:uid:gid]]",
					},
				},
			},
			{
				Name:      "sendlarge",
				Usage:     "send single large file to workload(s)",
				ArgsUsage: workloadArgsUsage,
				Action:    utils.ExitCoder(cmdWorkloadSendLarge),
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:  flagFile,
						Usage: "copy local file to workload, only can use single time. src_path:dst_path[:mode[:uid:gid]]",
					},
				},
			},
			{
				Name:      "dissociate",
				Usage:     "dissociate workload(s) from eru, return it resource but not remove it",
				ArgsUsage: workloadArgsUsage,
				Action:    utils.ExitCoder(cmdWorkloadDissociate),
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:  flagNode,
						Usage: "dissociate all workload(s) on node(s)",
					},
				},
			},
			{
				Name:      "realloc",
				Usage:     "realloc workload resource",
				ArgsUsage: "workloadID",
				Action:    utils.ExitCoder(cmdWorkloadRealloc),
				Flags: []cli.Flag{
					&cli.Float64Flag{
						Name:  flagCPURequest,
						Usage: "cpu request increment/decrement",
						Value: 0,
					},
					&cli.Float64Flag{
						Name:  flagCPULimit,
						Usage: "cpu limit increment/decrement",
						Value: 0,
					},
					&cli.Float64Flag{
						Name:  "cpu",
						Usage: "shortcut to set cpu-limit/request equally to this value",
						Value: 0,
					},
					&cli.StringFlag{
						Name:  flagMemoryRequest,
						Usage: "memory request increment/decrement, like -1M or 1G, support K, M, G, T",
					},
					&cli.StringFlag{
						Name:  flagMemoryLimit,
						Usage: "memory limit increment/decrement, like -1M or 1G, support K, M, G, T",
					},
					&cli.StringFlag{
						Name:  "memory",
						Usage: "shortcut to set memory-limit/request equally to this value",
					},
					&cli.StringFlag{
						Name:  flagVolumesRequest,
						Usage: `volumes request increment/decrement, like "AUTO:/data:rw:-1G,/tmp:/tmp"`,
					},
					&cli.StringFlag{
						Name:  flagVolumesLimit,
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
						Name:  flagStorageRequest,
						Usage: `storage request incr/decr, like "-1G"`,
					},
					&cli.StringFlag{
						Name:  flagStorageLimit,
						Usage: `storage limit incr/decr, like "-1G"`,
					},
					&cli.StringFlag{
						Name:  flagStorage,
						Usage: "shortcut to set storage-limit/request equally to this value",
					},
					&cli.StringFlag{
						Name:  utils.FlagExtraResources,
						Usage: "add extra resource requests",
						Value: "",
					},
				},
			},
			{
				Name:         "exec",
				Usage:        "run a command in a running workload",
				ArgsUsage:    "workloadID -- cmd1 cmd2 cmd3",
				StopOnNthArg: &stopOnFirstArg,
				Action:       utils.ExitCoder(cmdWorkloadExec),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "interactive",
						Aliases: []string{"i"},
						Usage:   "attach stdin to the command",
						Value:   false,
					},
					&cli.StringSliceFlag{
						Name:    flagEnv,
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
						Name:  flagPod,
						Usage: "where to replace",
					},
					&cli.StringFlag{
						Name:  flagEntry,
						Usage: "which entry",
					},
					&cli.StringFlag{
						Name:  flagImage,
						Usage: "which to replace",
					},
					&cli.StringSliceFlag{
						Name:  flagNode,
						Usage: "which node to replace",
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
						Name:  flagNetwork,
						Usage: "SDN name or host mode",
					},
					&cli.StringSliceFlag{
						Name:  flagEnv,
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
						Name:  flagFile,
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
						Name:  flagPod,
						Usage: "where to run",
					},
					&cli.StringFlag{
						Name:  flagEntry,
						Usage: "which entry",
					},
					&cli.StringFlag{
						Name:  flagImage,
						Usage: "which to run",
					},
					&cli.StringSliceFlag{
						Name:  flagNode,
						Usage: "which node to run",
					},
					&cli.IntFlag{
						Name:  "count",
						Usage: "how many",
						Value: 1,
					},
					&cli.StringFlag{
						Name:  flagNetwork,
						Usage: "SDN name or host mode",
						Value: "host",
					},
					&cli.Float64Flag{
						Name:  flagCPURequest,
						Usage: "how many cpu to request",
						Value: 0,
					},
					&cli.Float64Flag{
						Name:  flagCPULimit,
						Usage: "how many cpu to limit; can specify limit without request",
						Value: 1.0,
					},
					&cli.Float64Flag{
						Name:  "cpu",
						Usage: "shortcut for cpu-request/limit, set them equally to this value",
						Value: 1.0,
					},
					&cli.StringFlag{
						Name:  flagMemoryRequest,
						Usage: "how many memory to request like 1M or 1G, support K, M, G, T",
						Value: "",
					},
					&cli.StringFlag{
						Name:  flagMemoryLimit,
						Usage: "how many memory to limit like 1M or 1G, support K, M, G, T; can specify limit without request",
						Value: "512M",
					},
					&cli.StringFlag{
						Name:  "memory",
						Usage: "shortcut for memory-request/limit, set them equally to this value",
						Value: "512M",
					},
					&cli.StringFlag{
						Name:  flagStorageRequest,
						Usage: "how many storage to request quota like 1M or 1G, support K, M, G, T",
						Value: "",
					},
					&cli.StringFlag{
						Name:  flagStorageLimit,
						Usage: "how many storage to limit quota like 1M or 1G, support K, M, G, T; can specify limit without request",
						Value: "",
					},
					&cli.StringFlag{
						Name:  flagStorage,
						Usage: "shortcut for storage-request/limit, set them equally to this value",
						Value: "",
					},
					&cli.StringSliceFlag{
						Name:  flagEnv,
						Usage: "set env can use multiple times, e.g., GO111MODULE=on",
					},
					&cli.StringSliceFlag{
						Name:  "nodelabel",
						Usage: "filter nodes by labels",
					},
					&cli.StringFlag{
						Name:  "deploy-strategy",
						Usage: "deploy method auto/fill/each/global/drained/dummy",
						Value: strategy.Auto,
					},
					&cli.StringFlag{
						Name:  "user",
						Usage: "which user",
						Value: "root",
					},
					&cli.StringSliceFlag{
						Name:  flagFile,
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
					&cli.StringFlag{
						Name:  utils.FlagExtraResources,
						Usage: "add extra resource requests",
						Value: "",
					},
				},
			},
		},
	}
}
