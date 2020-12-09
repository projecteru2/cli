package pod

import (
	"context"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
)

type resourcePodOptions struct {
	client corepb.CoreRPCClient
	name   string
}

func (o *resourcePodOptions) run(ctx context.Context) error {
	resp, err := o.client.GetPodResource(ctx, &corepb.GetPodOptions{
		Name: o.name,
	})
	if err != nil {
		return err
	}

	describe.NodeResources(resp.NodesResource...)
	return nil
}

func cmdPodResource(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Pod name must be given")
	}

	o := &resourcePodOptions{
		client: client,
		name:   name,
	}
	return o.run(c.Context)
}
