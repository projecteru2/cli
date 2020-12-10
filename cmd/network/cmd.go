package network

import (
	"github.com/projecteru2/cli/cmd/utils"
	"github.com/urfave/cli/v2"
)

const (
	workloadArgsUsage = "workloadID(s)"
)

// Command exports network subommands
func Command() *cli.Command {
	return &cli.Command{
		Name:  "network",
		Usage: "network commands",
		Subcommands: []*cli.Command{
			{
				Name:      "connect",
				ArgsUsage: workloadArgsUsage,
				Usage:     "connect workloads to network",
				Action:    utils.ExitCoder(cmdNetworkConnect),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "network",
						Usage:    "network name",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "ipv4",
						Usage: "specify ipv4",
					},
					&cli.StringFlag{
						Name:  "ipv6",
						Usage: "specify ipv6",
					},
				},
			},
			{
				Name:      "disconnect",
				ArgsUsage: workloadArgsUsage,
				Usage:     "disconnect workloads to network",
				Action:    utils.ExitCoder(cmdNetworkDisconnect),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "network",
						Usage:    "network name",
						Required: true,
					},
				},
			},
		},
	}
}
