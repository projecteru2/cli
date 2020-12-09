package node

import (
	"context"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type removeNodeOptions struct {
	client corepb.CoreRPCClient
	name   string
}

func (o *removeNodeOptions) run(ctx context.Context) error {
	_, err := o.client.RemoveNode(ctx, &corepb.RemoveNodeOptions{
		Nodename: o.name,
	})
	if err != nil {
		return err
	}
	logrus.Infof("[RemoveNode] success")
	return nil
}

func cmdNodeRemove(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Node name must be given")
	}

	o := &removeNodeOptions{
		client: client,
		name:   name,
	}
	return o.run(c.Context)
}
