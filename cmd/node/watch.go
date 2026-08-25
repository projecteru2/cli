package node

import (
	"context"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
)

type watchNodeStatusOptions struct {
	client corepb.CoreRPCClient
}

func (o *watchNodeStatusOptions) run(ctx context.Context) error {
	resp, err := o.client.NodeStatusStream(ctx, &corepb.Empty{})
	if err != nil {
		return err
	}

	return utils.EachMessage(resp.Recv, func(m *corepb.NodeStatusStreamMessage) error {
		describe.NodeStatusMessage(ctx, m)
		return nil
	})
}

func cmdNodeWatchStatus(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	o := &watchNodeStatusOptions{
		client: client,
	}
	return o.run(ctx)
}
