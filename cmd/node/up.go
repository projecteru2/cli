package node

import (
	"context"
	"errors"

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

	name := cmd.Args().First()
	if name == "" {
		return errors.New("node name must be given")
	}

	o := &setNodeUpOptions{
		client: client,
		name:   name,
	}
	return o.run(ctx)
}
