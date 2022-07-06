package node

import (
	"context"
	"strings"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/juju/errors"
	"github.com/urfave/cli/v2"
)

type getNodeOptions struct {
	client  corepb.CoreRPCClient
	plugins []string
	name    string
}

func (o *getNodeOptions) run(ctx context.Context) error {
	node, err := o.client.GetNode(ctx, &corepb.GetNodeOptions{
		Nodename: o.name,
		Plugins:  o.plugins,
	})
	if err != nil {
		return err
	}

	describe.NodesWithInfo(describe.ToNodeChan(node), false)
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
	if c.IsSet("plugins") {
		o.plugins = strings.Split(c.String("plugins"), ",")
	}
	return o.run(c.Context)
}
