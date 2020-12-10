package pod

import (
	"github.com/projecteru2/cli/cmd/utils"
	"github.com/urfave/cli/v2"
)

const (
	podArgsUsage = "podname"
)

// Command exports pod subommands
func Command() *cli.Command {
	return &cli.Command{
		Name:  "pod",
		Usage: "pod commands",
		Subcommands: []*cli.Command{
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
			},
			{
				Name:      "nodes",
				Usage:     "list all nodes in one pod",
				ArgsUsage: podArgsUsage,
				Action:    utils.ExitCoder(cmdPodListNodes),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "all",
						Usage: "list all nodes or just living nodes",
						Value: false,
					},
					&cli.StringSliceFlag{
						Name:  "label",
						Usage: "labels to filter, e.g, a=1, b=2",
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
