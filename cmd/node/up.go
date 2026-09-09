package node

import (
	"context"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type setNodeUpOptions struct {
	client corepb.CoreRPCClient
	name   string
}

func (o *setNodeUpOptions) run(ctx context.Context) error {
	_, err := o.client.SetNode(ctx, &corepb.SetNodeOptions{
		Nodename: o.name,
		Bypass:   corepb.TriOpt_FALSE,
	})
	if err != nil {
		return err
	}
	log.WithFunc("node.setNodeUpOptions.run").Infof(ctx, "node %s up", o.name)
	return nil
}

func cmdNodeSetUp(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	o := &setNodeUpOptions{
		client: client,
		name:   cmd.StringArgs(argNode)[0],
	}
	return o.run(ctx)
}
