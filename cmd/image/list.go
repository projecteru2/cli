package image

import (
	"context"
	"errors"
	"fmt"
	"io"

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
	for {
		msg, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Printf("Build failed: %s\n", err.Error())
			return err
		}

		if len(msg.Err) > 1 {
			fmt.Printf("Build failed: %s\n", msg.Err)
			return cli.Exit(msg.Err, -1)
		}

		msgs = append(msgs, msg)
	}

	describe.Images(msgs...)
	return nil
}

func cmdImageList(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	opts, err := generateListOptions(cmd)
	if err != nil {
		return err
	}

	o := &listImageOptions{
		client: client,
		opts:   opts,
	}
	return o.run(ctx)
}

func generateListOptions(cmd *cli.Command) (*corepb.ListImageOptions, error) {
	filter := cmd.String("filter")
	podname := cmd.String("pod")
	nodename := cmd.StringSlice("node")
	if len(nodename) < 1 && len(podname) < 1 {
		return nil, errors.New("podname or nodenames should be given")
	}

	return &corepb.ListImageOptions{
		Podname:   podname,
		Nodenames: nodename,
		Filter:    filter,
	}, nil
}
