package core

import (
	"context"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
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

func cmdCoreInfo(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	o := &coreInfoOptions{
		client: client,
	}
	return o.run(c.Context)
}
