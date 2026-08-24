package node

import (
	"context"
	"errors"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type removeNodeOptions struct {
	client corepb.CoreRPCClient
	name   string
}

func (o *removeNodeOptions) run(ctx context.Context) error {
	_, err := o.client.RemoveNode(ctx, &corepb.RemoveNodeOptions{
		Nodename: o.name,
	})
	if err != nil {
		return err
	}
	log.WithFunc("node.removeNodeOptions.run").Infof(ctx, "remove node %s success", o.name)
	return nil
}

func cmdNodeRemove(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	name := cmd.Args().First()
	if name == "" {
		return errors.New("Node name must be given")
	}

	o := &removeNodeOptions{
		client: client,
		name:   name,
	}
	return o.run(ctx)
}
