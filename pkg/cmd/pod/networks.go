package pod

import (
	"context"
	"strings"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/pkg/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type listPodNetworksOptions struct {
	client corepb.CoreRPCClient
	name   string
	driver string
}

func (o *listPodNetworksOptions) run(ctx context.Context) error {
	resp, err := o.client.ListNetworks(ctx, &corepb.ListNetworkOptions{
		Podname: o.name,
		Driver:  o.driver,
	})
	if err != nil {
		return err
	}

	for _, network := range resp.GetNetworks() {
		logrus.Infof("Name: %s, Subnets: %s", network.GetName(), strings.Join(network.GetSubnets(), ","))
	}
	return nil
}

func cmdPodListNetworks(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Pod name must be given")
	}

	o := &listPodNetworksOptions{
		client: client,
		name:   name,
		driver: c.String("driver"),
	}
	return o.run(c.Context)
}
