package node

import (
	"context"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
)

type getNodeOptions struct {
	client corepb.CoreRPCClient
	name   string
}

func (o *getNodeOptions) run(ctx context.Context) error {
	node, err := o.client.GetNode(ctx, &corepb.GetNodeOptions{
		Nodename: o.name,
	})
	if err != nil {
		return err
	}

	describe.Nodes(node)
	return nil
}

func cmdNodeGet(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Node name must be given")
	}

	o := &getNodeOptions{
		client: client,
		name:   name,
	}
	return o.run(c.Context)
}
