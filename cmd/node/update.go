package node

import (
	"context"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type updateNodeOptions struct {
	client   corepb.CoreRPCClient
	name     string
	endpoint string
}

func (o *updateNodeOptions) run(ctx context.Context) error {
	_, err := o.client.SetNode(ctx, &corepb.SetNodeOptions{
		Nodename: o.name,
		Endpoint: o.endpoint,
	})
	if err != nil {
		return err
	}
	logrus.Infof("[UpdateNode] success")
	return nil
}

func cmdNodeUpdate(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Node name must be given")
	}

	endpoint := c.String("endpoint")

	o := &updateNodeOptions{
		client:   client,
		name:     name,
		endpoint: endpoint,
	}
	return o.run(c.Context)
}
