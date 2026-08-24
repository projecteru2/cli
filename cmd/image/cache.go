package image

import (
	"context"
	"errors"
	"io"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type cacheImageOptions struct {
	client    corepb.CoreRPCClient
	images    []string
	step      int32
	podname   string
	nodenames []string
}

func (o *cacheImageOptions) run(ctx context.Context) error {
	logger := log.WithFunc("image.cacheImageOptions.run")
	opts := &corepb.CacheImageOptions{
		Images:    o.images,
		Podname:   o.podname,
		Nodenames: o.nodenames,
	}
	resp, err := o.client.CacheImage(ctx, opts)
	if err != nil {
		return err
	}

	for {
		msg, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		if msg.Success {
			logger.Infof(ctx, "cache image %s on %s success", msg.Image, msg.Nodename)
		} else {
			logger.Warnf(ctx, "cache image %s on %s failed", msg.Image, msg.Nodename)
		}
	}
	return nil
}

func cmdImageCache(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	images := cmd.Args().Slice()
	if len(images) == 0 {
		return errors.New("images must be specified")
	}

	o := &cacheImageOptions{
		client:    client,
		images:    images,
		step:      int32(cmd.Int("concurrent")), //nolint:gosec
		podname:   cmd.String("pod"),
		nodenames: cmd.StringSlice("node"),
	}
	return o.run(ctx)
}
