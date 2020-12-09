package node

import (
	"context"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
)

type nodeResourceOptions struct {
	client corepb.CoreRPCClient
	name   string
	fix    bool
}

func (o *nodeResourceOptions) run(ctx context.Context) error {
	resource, err := o.client.GetNodeResource(ctx, &corepb.GetNodeResourceOptions{
		Opts: &corepb.GetNodeOptions{Nodename: o.name},
		Fix:  o.fix,
	},
	)
	if err != nil {
		return err
	}

	describe.NodeResources(resource)
	return nil
}

func cmdNodeResource(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Node name must be given")
	}

	o := &nodeResourceOptions{
		client: client,
		name:   name,
		fix:    c.Bool("fix"),
	}
	return o.run(c.Context)
}
