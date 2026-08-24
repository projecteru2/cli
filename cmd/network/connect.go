package network

import (
	"context"
	"errors"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type connectNetworkOptions struct {
	client  corepb.CoreRPCClient
	ids     []string
	network string
	ipv4    string
	ipv6    string
}

func (o *connectNetworkOptions) run(ctx context.Context) error {
	logger := log.WithFunc("network.connectNetworkOptions.run")
	for _, id := range o.ids {
		resp, err := o.client.ConnectNetwork(ctx, &corepb.ConnectNetworkOptions{
			Network: o.network,
			Target:  id,
			Ipv4:    o.ipv4,
			Ipv6:    o.ipv6,
		})
		if err != nil {
			logger.Warnf(ctx, "connect %s to network %s failed: %v", id, o.network, err)
		} else {
			logger.Infof(ctx, "connect %s at %v", id, resp.Subnets)
		}
	}
	return nil
}

func cmdNetworkConnect(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	ids := cmd.Args().Slice()
	if len(ids) == 0 {
		return errors.New("Workload ID(s) must be specified")
	}

	network := cmd.String("network")
	if network == "" {
		return errors.New("Network must be specified")
	}

	o := &connectNetworkOptions{
		client:  client,
		ids:     ids,
		network: network,
		ipv4:    cmd.String("ipv4"),
		ipv6:    cmd.String("ipv6"),
	}
	return o.run(ctx)
}
