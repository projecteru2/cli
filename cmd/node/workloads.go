package node

import (
	"context"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
)

type listNodeWorkloadsOptions struct {
	client corepb.CoreRPCClient
	name   string
	labels map[string]string
}

func (o *listNodeWorkloadsOptions) run(ctx context.Context) error {
	resp, err := o.client.ListNodeWorkloads(ctx, &corepb.GetNodeOptions{
		Nodename: o.name,
		Labels:   o.labels,
	})
	if err != nil {
		return err
	}

	describe.Workloads(resp.Workloads...)
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
		labels: utils.SplitEquality(c.StringSlice("label")),
	}
	return o.run(c.Context)
}
