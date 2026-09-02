package image

import (
	"context"
	"fmt"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
)

type listImageOptions struct {
	client corepb.CoreRPCClient
	opts   *corepb.ListImageOptions
}

func (o *listImageOptions) run(ctx context.Context) error {
	resp, err := o.client.ListImage(ctx, o.opts)
	if err != nil {
		return err
	}

	msgs := []*corepb.ListImageMessage{}
	err = utils.EachMessage(resp.Recv, func(msg *corepb.ListImageMessage) error {
		if msg.Err != "" {
			return fmt.Errorf("list images on %s: %s", msg.Nodename, msg.Err)
		}
		msgs = append(msgs, msg)
		return nil
	})

	describe.Images(msgs...)
	return err
}

func cmdImageList(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	o := &listImageOptions{
		client: client,
		opts:   generateListOptions(cmd),
	}
	return o.run(ctx)
}

func generateListOptions(cmd *cli.Command) *corepb.ListImageOptions {
	return &corepb.ListImageOptions{
		Podname:   cmd.String(flagPod),
		Nodenames: cmd.StringSlice(flagNode),
		Filter:    cmd.String("filter"),
	}
}
