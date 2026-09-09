package node

import (
	"context"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
)

type nodeResourceOptions struct {
	client corepb.CoreRPCClient
	name   string
	fix    bool
}

func (o *nodeResourceOptions) run(ctx context.Context) error {
	resource, err := o.client.GetNodeResource(ctx, &corepb.GetNodeResourceOptions{
		Opts: &corepb.GetNodeOptions{Nodename: o.name},
		Fix:  o.fix,
	},
	)
	if err != nil {
		return err
	}

	describe.NodeResource(ctx, resource)
	return nil
}

func cmdNodeResource(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	o := &nodeResourceOptions{
		client: client,
		name:   cmd.StringArgs(argNode)[0],
		fix:    cmd.Bool("fix"),
	}
	return o.run(ctx)
}
