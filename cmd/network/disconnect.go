package network

import (
	"context"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type disconnectNetworkOptions struct {
	client  corepb.CoreRPCClient
	ids     []string
	network string
}

func (o *disconnectNetworkOptions) run(ctx context.Context) error {
	for _, id := range o.ids {
		if _, err := o.client.DisconnectNetwork(ctx, &corepb.DisconnectNetworkOptions{
			Network: o.network,
			Target:  id,
		}); err != nil {
			logrus.Warnf("[disConnectToNetwork] Disconnect %s to network %s failed", id, o.network)
		} else {
			logrus.Infof("[disConnectToNetwork] Disconnect %s success", id)
		}
	}
	return nil
}

func cmdNetworkDisconnect(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	ids := c.Args().Slice()
	if len(ids) == 0 {
		return errors.New("Workload ID(s) must be specified")
	}

	network := c.String("network")
	if network == "" {
		return errors.New("Network must be specified")
	}

	o := &disconnectNetworkOptions{
		client:  client,
		ids:     ids,
		network: network,
	}
	return o.run(c.Context)
}
