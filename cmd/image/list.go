package image

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	"github.com/urfave/cli/v2"

	corepb "github.com/projecteru2/core/rpc/gen"
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
		if err == io.EOF {
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

func cmdImageList(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	opts, err := generateListOptions(c)
	if err != nil {
		return err
	}

	o := &listImageOptions{
		client: client,
		opts:   opts,
	}
	return o.run(c.Context)
}

func generateListOptions(c *cli.Context) (*corepb.ListImageOptions, error) {
	filter := c.String("filter")
	podname := c.String("podname")
	nodename := c.StringSlice("nodename")
	if len(nodename) < 1 && len(podname) < 1 {
		return nil, errors.New("[List] podname or nodenames should be given")
	}

	return &corepb.ListImageOptions{
		Podname:   podname,
		Nodenames: nodename,
		Filter:    filter,
	}, nil
}
