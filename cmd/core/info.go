package core

import (
	"context"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
)

type coreInfoOptions struct {
	client corepb.CoreRPCClient
}

func (o *coreInfoOptions) run(ctx context.Context) error {
	info, err := o.client.Info(ctx, &corepb.Empty{})
	if err != nil {
		return err
	}

	describe.Core(info)
	return nil
}

func cmdCoreInfo(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	o := &coreInfoOptions{
		client: client,
	}
	return o.run(ctx)
}
