package node

import (
	"github.com/projecteru2/cli/cmd/utils"

	"github.com/urfave/cli/v2"
)

const (
	nodeArgsUsage = "node name"
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
				Flags:     []cli.Flag{},
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
				Name:  "set-status",
				Usage: "set status of node, used for heartbeat",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "ttl",
						Usage: "status ttl for node",
						Value: 180,
					},
					&cli.IntFlag{
						Name:  "interval",
						Usage: "if given, will set status every INTERVAL seconds",
						Value: 0,
					},
				},
				ArgsUsage: nodeArgsUsage,
				Action:    utils.ExitCoder(cmdNodeSetStatus),
			},
			{
				Name:   "watch-status",
				Usage:  "watch status of node, used for heartbeat",
				Action: utils.ExitCoder(cmdNodeWatchStatus),
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
				Aliases:   []string{"update"},
				Usage:     "set node resource",
				ArgsUsage: nodeArgsUsage,
				Action:    utils.ExitCoder(cmdNodeSet),
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "mark-workloads-down",
						Usage: "mark workloads down",
					},
					&cli.StringFlag{
						Name:  "cpu",
						Usage: "cpu value in string, e.g. 0:100,1:200,3:50",
					},
					&cli.StringFlag{
						Name: "memory",
						Usage: `memory, unit can be K/M/G/T, 
                          when using --delta flag, this can be a negtive number indicating how much to add to the current value, 
                          e.g. --memory -10G --delta, means memory will be the current value - 10`,
					},
					&cli.StringSliceFlag{
						Name: "numa-cpu",
						Usage: `numa cpu list, can be set multiple times, the index will be the numa node ID.
                     e.g. --numa-cpu 0,1,2,3 --numa-cpu 4,5,6,7 means cpu 0,1,2,3 are bound to node ID 0, cpu 4,5,6,7 are bound to node ID 1`,
					},
					&cli.StringSliceFlag{
						Name: "numa-memory",
						Usage: `numa memory values, unit can be K/M/G/T, 
                        when using --delta flag, this can be a negtive number indicating how much to add to the current value, 
                        e.g. --numa-memory -10G --delta, means the value will be current value - 10
                        this value can be set multiple times, the index will be the numa node ID,
                        e.g. --numa-memory 10G --numa-memory 15G, means node ID 0 will be 10GB, node ID 1 will be 15GB`,
					},
					&cli.StringFlag{
						Name: "storage",
						Usage: `storage, unit can be K/M/G/T,
					            when using --delta flag, this can be a negtive number indicating how much to add to the current value,
					            e.g. --storage -10G --delta, means storage will be the current value - 10`,
					},
					&cli.StringSliceFlag{
						Name: "volume",
						Usage: `volume value in string, can set multiple times. e.g. "--volume /data:100G",
					            when using --delta flag, this can be a negative number indicating how much to add to the current value,
					            e.g. --volume /data0:-10G --volume /data1:20G, means /data0 will be subtract 10G and /data1 will be added 20G`,
					},
					&cli.StringSliceFlag{
						Name: "disk",
						Usage: `disk value in string, format: device:mounts:read-iops:write-iops:read-bps:write-bps
								e.g. --disk /dev/sda1:/data0:100:100:100M:100M
								when using --delta flag, this can be a negative number indicating how much to add to the current value`,
					},
					&cli.StringFlag{
						Name: "rm-disk",
						Usage: `remove disks, e.g. --rm-disk /dev/vda,/dev/vdb
								rm-disk is not supported in delta mode`,
					},
					&cli.StringSliceFlag{
						Name:  "label",
						Usage: "label for the node, can set multiple times, e.g. --label a=1 --label b=2",
					},
					&cli.BoolFlag{
						Name:  "delta",
						Usage: "delta flag for settings, when set, all values will be relative to the current values, refer to each option for details",
					},
					&cli.StringFlag{
						Name:  "endpoint",
						Usage: "update node endpoint",
					},
					&cli.StringFlag{
						Name:  "ca",
						Usage: "ca file, like /etc/docker/tls/ca.crt",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "cert",
						Usage: "cert file, like /etc/docker/tls/client.crt",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "key",
						Usage: "key file, like /etc/docker/tls/client.key",
						Value: "",
					},
				},
			},
			{
				Name:      "add",
				Usage:     "add node",
				ArgsUsage: "pod name",
				Action:    utils.ExitCoder(cmdNodeAdd),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "nodename",
						Usage:   "name of this node, use `hostname` as default",
						EnvVars: []string{"HOSTNAME"},
						Value:   utils.GetHostname(),
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
						Name:  "volume",
						Usage: `device volumes, can set multiple times. e.g. "--volume /data:100G" `,
					},
					&cli.StringSliceFlag{
						Name: "disk",
						Usage: `disk value in string, format: device:mounts:read-iops:write-iops:read-bps:write-bps
										e.g. --disk /dev/sda1:/data0:100:100:100M:100M`,
					},
				},
			},
		},
	}
}
