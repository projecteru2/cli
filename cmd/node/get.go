package node

import (
	"context"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
)

type getNodeOptions struct {
	client corepb.CoreRPCClient
	name   string
}

func (o *getNodeOptions) run(ctx context.Context) error {
	node, err := o.client.GetNode(ctx, &corepb.GetNodeOptions{
		Nodename: o.name,
	})
	if err != nil {
		return err
	}

	describe.Nodes(true, node)
	return nil
}

func cmdNodeGet(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	o := &getNodeOptions{
		client: client,
		name:   cmd.StringArgs(argNode)[0],
	}
	return o.run(ctx)
}
