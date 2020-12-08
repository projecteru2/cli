package node

import (
	"context"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/pkg/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type listNodeWorkloadsOptions struct {
	client corepb.CoreRPCClient
	name   string
}

func (o *listNodeWorkloadsOptions) run(ctx context.Context) error {
	resp, err := o.client.ListNodeWorkloads(ctx, &corepb.GetNodeOptions{
		Nodename: o.name,
	})
	if err != nil {
		return err
	}

	logrus.Info(resp.Workloads)
	return nil
}

func cmdNodeListWorkloads(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Node name must be given")
	}

	o := &listNodeWorkloadsOptions{
		client: client,
		name:   name,
	}
	return o.run(c.Context)
}
