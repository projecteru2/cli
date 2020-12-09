package node

import (
	"context"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type setNodeUpOptions struct {
	client corepb.CoreRPCClient
	name   string
}

func (o *setNodeUpOptions) run(ctx context.Context) error {
	_, err := o.client.SetNode(ctx, &corepb.SetNodeOptions{
		Nodename:  o.name,
		StatusOpt: corepb.TriOpt_TRUE,
	})
	if err != nil {
		return err
	}
	logrus.Infof("[SetNode] node %s up", o.name)
	return nil
}

func cmdNodeSetUp(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Node name must be given")
	}

	o := &setNodeUpOptions{
		client: client,
		name:   name,
	}
	return o.run(c.Context)
}
