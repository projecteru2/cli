package network

import (
	"context"
	"errors"
	"fmt"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"
)

type disconnectNetworkOptions struct {
	client  corepb.CoreRPCClient
	ids     []string
	network string
}

func (o *disconnectNetworkOptions) run(ctx context.Context) error {
	logger := log.WithFunc("network.disconnectNetworkOptions.run")
	var errs error
	for _, id := range o.ids {
		if _, err := o.client.DisconnectNetwork(ctx, &corepb.DisconnectNetworkOptions{
			Network: o.network,
			Target:  id,
		}); err != nil {
			errs = errors.Join(errs, fmt.Errorf("disconnect %s from network %s: %w", id, o.network, err))
			continue
		}
		logger.Infof(ctx, "disconnect %s success", id)
	}
	return errs
}

func cmdNetworkDisconnect(ctx context.Context, cmd *cli.Command) error {
	client, ids, network, err := networkTarget(ctx, cmd)
	if err != nil {
		return err
	}

	o := &disconnectNetworkOptions{
		client:  client,
		ids:     ids,
		network: network,
	}
	return o.run(ctx)
}
