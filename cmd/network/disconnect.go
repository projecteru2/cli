package network

import (
	"context"
	"errors"

	"github.com/urfave/cli/v3"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/projecteru2/cli/cmd/utils"
)

type disconnectNetworkOptions struct {
	client  corepb.CoreRPCClient
	ids     []string
	network string
}

func (o *disconnectNetworkOptions) run(ctx context.Context) error {
	logger := log.WithFunc("network.disconnectNetworkOptions.run")
	for _, id := range o.ids {
		if _, err := o.client.DisconnectNetwork(ctx, &corepb.DisconnectNetworkOptions{
			Network: o.network,
			Target:  id,
		}); err != nil {
			logger.Warnf(ctx, "disconnect %s from network %s failed: %v", id, o.network, err)
		} else {
			logger.Infof(ctx, "disconnect %s success", id)
		}
	}
	return nil
}

func cmdNetworkDisconnect(ctx context.Context, cmd *cli.Command) error {
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

	o := &disconnectNetworkOptions{
		client:  client,
		ids:     ids,
		network: network,
	}
	return o.run(ctx)
}
