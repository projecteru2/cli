package node

import (
	"context"
	"io"

	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"

	"github.com/urfave/cli/v3"
)

type watchNodeStatusOptions struct {
	client corepb.CoreRPCClient
}

func (o *watchNodeStatusOptions) run(ctx context.Context) error {
	resp, err := o.client.NodeStatusStream(ctx, &corepb.Empty{})
	if err != nil {
		return err
	}

	for {
		m, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		describe.NodeStatusMessage(ctx, m)
	}
	return nil
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
