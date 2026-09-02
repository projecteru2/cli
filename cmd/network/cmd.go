package network

import (
	"context"
	"errors"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

const (
	workloadArgsUsage = "workloadID(s)"

	flagNetwork = "network"
)

// Command returns the network command tree.
func Command() *cli.Command {
	return &cli.Command{
		Name:  "network",
		Usage: "network commands",
		Commands: []*cli.Command{
			{
				Name:      "connect",
				ArgsUsage: workloadArgsUsage,
				Usage:     "connect workloads to network",
				Action:    utils.ExitCoder(cmdNetworkConnect),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     flagNetwork,
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
				Usage:     "disconnect workloads from network",
				Action:    utils.ExitCoder(cmdNetworkDisconnect),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     flagNetwork,
						Usage:    "network name",
						Required: true,
					},
				},
			},
		},
	}
}

func networkTarget(ctx context.Context, cmd *cli.Command) (corepb.CoreRPCClient, []string, string, error) {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return nil, nil, "", err
	}

	ids := cmd.Args().Slice()
	if len(ids) == 0 {
		return nil, nil, "", errors.New("workload id(s) must be specified")
	}

	network := cmd.String(flagNetwork)
	if network == "" {
		return nil, nil, "", errors.New("network must be specified")
	}
	return client, ids, network, nil
}
