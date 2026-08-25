package pod

import (
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

const (
	podArgsUsage = "pod name"

	up   = "up"
	down = "down"
	all  = "all"

	flagCPU     = "cpu"
	flagMemory  = "memory"
	flagStorage = "storage"
)

// Command returns the pod command tree.
func Command() *cli.Command {
	return &cli.Command{
		Name:  "pod",
		Usage: "pod commands",
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "list all pods",
				Action: utils.ExitCoder(cmdPodList),
			},
			{
				Name:      "add",
				Usage:     "add new pod",
				ArgsUsage: podArgsUsage,
				Action:    utils.ExitCoder(cmdPodAdd),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "desc",
						Usage: "description of pod",
						Value: "",
					},
				},
			},
			{
				Name:      "remove",
				Usage:     "remove pod",
				ArgsUsage: podArgsUsage,
				Action:    utils.ExitCoder(cmdPodRemove),
			},
			{
				Name:      "resource",
				Usage:     "pod resource usage",
				ArgsUsage: podArgsUsage,
				Action:    utils.ExitCoder(cmdPodResource),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "filter",
						Aliases: []string{"f"},
						Usage:   "filter resource value, can be cpu/memory/storage/volume </<=/>/>=/== 40% or 0.4",
						Value:   all,
					},
					&cli.BoolFlag{
						Name:  "stream",
						Usage: "fetch streaming data",
					},
				},
			},
			{
				Name:      "capacity",
				Usage:     "pod remained capacity",
				ArgsUsage: podArgsUsage,
				Action:    utils.ExitCoder(cmdPodCapacity),
				Flags: []cli.Flag{
					&cli.Float64Flag{
						Name:     flagCPU,
						Aliases:  []string{"c"},
						Usage:    "how many cpu to occupy",
						Required: true,
					},
					&cli.StringFlag{
						Name:     flagMemory,
						Aliases:  []string{"m", "mem"},
						Usage:    "how much memory to occupy like 1M or 1G, support K, M, G, T",
						Required: true,
					},
					&cli.StringFlag{
						Name:     flagStorage,
						Aliases:  []string{"s"},
						Usage:    "how much storage to occupy like 1M or 1G, support K, M, G, T",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "cpu-bind",
						Usage: "bind cpu or not",
						Value: false,
					},
					&cli.StringSliceFlag{
						Name:     "node",
						Aliases:  []string{"n"},
						Usage:    "Specified the node(s) should join into the calculation. Could be specified multiple times with different names",
						Required: false,
					},
					&cli.StringFlag{
						Name:  "extra-resources",
						Usage: "add extra resource requests",
						Value: "",
					},
				},
			},
			{
				Name:      "nodes",
				Usage:     "list all nodes in one pod",
				ArgsUsage: podArgsUsage,
				Action:    utils.ExitCoder(cmdPodListNodes),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "filter",
						Aliases: []string{"f"},
						Usage:   "filter node status, can be up/down/all",
						Value:   all,
					},
					&cli.StringSliceFlag{
						Name:  "label",
						Usage: "labels to filter, e.g, a=1, b=2",
					},
					&cli.IntFlag{
						Name:  "timeout",
						Usage: "timeout in second, default value is 10",
						Value: 10,
					},
					&cli.BoolFlag{
						Name:  "show-info",
						Usage: "show node info",
					},
					&cli.BoolFlag{
						Name:  "stream",
						Usage: "fetch streaming data",
					},
				},
			},
			{
				Name:      "networks",
				Usage:     "list all networks in one pod",
				ArgsUsage: podArgsUsage,
				Action:    utils.ExitCoder(cmdPodListNetworks),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "driver",
						Usage: "filter driver",
					},
				},
			},
		},
	}
}
