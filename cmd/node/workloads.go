package node

import (
	"context"
	"errors"

	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"

	"github.com/urfave/cli/v3"
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

func cmdNodeListWorkloads(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	name := cmd.Args().First()
	if name == "" {
		return errors.New("Node name must be given")
	}

	o := &listNodeWorkloadsOptions{
		client: client,
		name:   name,
		labels: utils.SplitEquality(cmd.StringSlice("label")),
	}
	return o.run(ctx)
}
