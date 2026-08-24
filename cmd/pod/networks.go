package pod

import (
	"context"
	"errors"

	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"

	"github.com/urfave/cli/v3"
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

	describe.Networks(resp.GetNetworks()...)
	return nil
}

func cmdPodListNetworks(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	name := cmd.Args().First()
	if name == "" {
		return errors.New("Pod name must be given")
	}

	o := &listPodNetworksOptions{
		client: client,
		name:   name,
		driver: cmd.String("driver"),
	}
	return o.run(ctx)
}
