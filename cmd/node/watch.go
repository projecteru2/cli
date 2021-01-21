package node

import (
	"context"
	"io"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
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

		describe.NodeStatusMessage(m)
	}
	return nil
}

func cmdNodeWatchStatus(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	o := &watchNodeStatusOptions{
		client: client,
	}
	return o.run(c.Context)
}
