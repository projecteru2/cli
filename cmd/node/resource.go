package node

import (
	"context"
	"errors"

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

	describe.NodeResources(ctx, describe.ToNodeResourceChan(resource), false)
	return nil
}

func cmdNodeResource(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	name := cmd.Args().First()
	if name == "" {
		return errors.New("Node name must be given")
	}

	o := &nodeResourceOptions{
		client: client,
		name:   name,
		fix:    cmd.Bool("fix"),
	}
	return o.run(ctx)
}
