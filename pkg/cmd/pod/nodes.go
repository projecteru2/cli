package pod

import (
	"context"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/pkg/cmd/utils"
	"github.com/projecteru2/cli/pkg/describe"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
)

type listPodNodesOptions struct {
	client corepb.CoreRPCClient
	name   string
	all    bool
}

func (o *listPodNodesOptions) run(ctx context.Context) error {
	resp, err := o.client.ListPodNodes(ctx, &corepb.ListNodesOptions{
		Podname: o.name,
		All:     o.all,
	})
	if err != nil {
		return err
	}

	describe.DescribeNodes(resp.GetNodes()...)
	return nil
}

func cmdPodListNodes(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Pod name must be given")
	}

	o := &listPodNodesOptions{
		client: client,
		name:   name,
		all:    c.Bool("all"),
	}
	return o.run(c.Context)
}
