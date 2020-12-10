package pod

import (
	"context"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
)

type listPodNodesOptions struct {
	client corepb.CoreRPCClient
	name   string
	all    bool
	labels map[string]string
}

func (o *listPodNodesOptions) run(ctx context.Context) error {
	resp, err := o.client.ListPodNodes(ctx, &corepb.ListNodesOptions{
		Podname: o.name,
		All:     o.all,
		Labels:  o.labels,
	})
	if err != nil {
		return err
	}

	describe.Nodes(resp.GetNodes()...)
	return nil
}

func cmdPodListNodes(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	o := &listPodNodesOptions{
		client: client,
		name:   c.Args().First(),
		all:    c.Bool("all"),
		labels: utils.SplitEquality(c.StringSlice("label")),
	}
	return o.run(c.Context)
}
