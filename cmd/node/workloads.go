package node

import (
	"context"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
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

	o := &listNodeWorkloadsOptions{
		client: client,
		name:   cmd.StringArgs(argNode)[0],
		labels: utils.SplitEquality(cmd.StringSlice(flagLabel)),
	}
	return o.run(ctx)
}
