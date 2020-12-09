package network

import (
	"context"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type connectNetworkOptions struct {
	client  corepb.CoreRPCClient
	ids     []string
	network string
	ipv4    string
	ipv6    string
}

func (o *connectNetworkOptions) run(ctx context.Context) error {
	for _, id := range o.ids {
		resp, err := o.client.ConnectNetwork(ctx, &corepb.ConnectNetworkOptions{
			Network: o.network,
			Target:  id,
			Ipv4:    o.ipv4,
			Ipv6:    o.ipv6,
		})
		if err != nil {
			logrus.Warnf("[connectToNetwork] Connect %s to network %s failed", id, o.network)
		} else {
			logrus.Infof("[connectToNetwork] Connect %s at %v", id, resp.Subnets)
		}
	}
	return nil
}

func cmdNetworkConnect(c *cli.Context) error {
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

	o := &connectNetworkOptions{
		client:  client,
		ids:     ids,
		network: network,
		ipv4:    c.String("ipv4"),
		ipv6:    c.String("ipv6"),
	}
	return o.run(c.Context)
}
