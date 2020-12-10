package node

import (
	"github.com/projecteru2/cli/cmd/utils"
	"github.com/urfave/cli/v2"
)

const (
	nodeArgsUsage = "nodename"
)

// Command exports node subommands
func Command() *cli.Command {
	return &cli.Command{
		Name:  "node",
		Usage: "node commands",
		Subcommands: []*cli.Command{
			{
				Name:      "get",
				Usage:     "get a node",
				ArgsUsage: nodeArgsUsage,
				Action:    utils.ExitCoder(cmdNodeGet),
			},
			{
				Name:      "remove",
				Usage:     "remove a node",
				ArgsUsage: nodeArgsUsage,
				Action:    utils.ExitCoder(cmdNodeRemove),
			},
			{
				Name:  "workloads",
				Usage: "list node workloads",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:  "label",
						Usage: "labels to filter, e.g, a=1, b=2",
					},
				},
				Aliases:   []string{"containers"},
				ArgsUsage: nodeArgsUsage,
				Action:    utils.ExitCoder(cmdNodeListWorkloads),
			},
			{
				Name:      "up",
				Usage:     "set node up",
				ArgsUsage: nodeArgsUsage,
				Action:    utils.ExitCoder(cmdNodeSetUp),
			},
			{
				Name:  "down",
				Usage: "set node down",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "check",
						Usage: "check node workloads are online or not",
					},
					&cli.IntFlag{
						Name:  "check-timeout",
						Usage: "check node timeout",
						Value: 20,
					},
				},
				ArgsUsage: nodeArgsUsage,
				Action:    utils.ExitCoder(cmdNodeSetDown),
			},
			{
				Name:      "resource",
				Usage:     "check node resource",
				ArgsUsage: nodeArgsUsage,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "fix",
						Usage: "fix node resource diff",
					},
				},
				Action: utils.ExitCoder(cmdNodeResource),
			},
			{
				Name:      "set",
				Usage:     "set node resource",
				ArgsUsage: nodeArgsUsage,
				Action:    utils.ExitCoder(cmdNodeSet),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "mark-workloads-down",
						Usage: "mark workloads down",
					},
					&cli.StringFlag{
						Name:  "delta-memory",
						Usage: "memory changes like -1M or 1G, support K, M, G, T",
					},
					&cli.StringFlag{
						Name:  "delta-storage",
						Usage: "storage changes like -1M or 1G, support K, M, G, T",
					},
					&cli.StringFlag{
						Name:  "delta-cpu",
						Usage: "cpu changes in string, like 0:100,1:200,3:50",
					},
					&cli.StringSliceFlag{
						Name:  "delta-numa-memory",
						Usage: "numa memory changes, can set multiple times, like -1M or 1G, support K, M, G, T",
					},
					&cli.StringFlag{
						Name:  "delta-volume",
						Usage: `volume changed in string, like "/data0:-1G,/data1:1G"`,
					},
					&cli.StringSliceFlag{
						Name:  "numa-cpu",
						Usage: "numa cpu list, can set multiple times, use comma separated",
					},
					&cli.StringSliceFlag{
						Name:  "label",
						Usage: "add label for node, like a=1 b=2, can set multiple times",
					},
				},
			},
			{
				Name:      "add",
				Usage:     "add node",
				ArgsUsage: "podname",
				Action:    utils.ExitCoder(cmdNodeAdd),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "nodename",
						Usage:   "name of this node, use `hostname` as default",
						EnvVars: []string{"HOSTNAME"},
						Value:   "",
					},
					&cli.StringFlag{
						Name:  "endpoint",
						Usage: "endpoint of docker server",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "ca",
						Usage: "ca file of docker server, like /etc/docker/tls/ca.crt",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "cert",
						Usage: "cert file of docker server, like /etc/docker/tls/client.crt",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "key",
						Usage: "key file of docker server, like /etc/docker/tls/client.key",
						Value: "",
					},
					&cli.IntFlag{
						Name:        "cpu",
						Usage:       "cpu count",
						DefaultText: "total cpu",
					},
					&cli.IntFlag{
						Name:        "share",
						Usage:       "share count",
						DefaultText: "defined in core",
					},
					&cli.StringFlag{
						Name:  "memory",
						Usage: "memory like -1M or 1G, support K, M, G, T",
					},
					&cli.StringFlag{
						Name:  "storage",
						Usage: "storage -1M or 1G, support K, M, G, T",
					},
					&cli.StringSliceFlag{
						Name:  "label",
						Usage: "add label for node, like a=1 b=2, can set multiple times",
					},
					&cli.StringSliceFlag{
						Name:  "numa-cpu",
						Usage: "numa cpu list, can set multiple times, use comma separated",
					},
					&cli.StringSliceFlag{
						Name:  "numa-memory",
						Usage: "numa memory, can set multiple times. if not set, it will count numa-cpu groups, and divided by total memory",
					},
					&cli.StringSliceFlag{
						Name:  "volumes",
						Usage: `device volumes, can set multiple times. e.g. "--volumes /data:100G" `,
					},
				},
			},
		},
	}
}
