package image

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type cacheImageOptions struct {
	client    corepb.CoreRPCClient
	images    []string
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

	var errs error
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
			errs = errors.Join(errs, fmt.Errorf("cache image %s on %s: %s", msg.Image, msg.Nodename, msg.Message))
		}
	}
	return errs
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
		podname:   cmd.String(flagPod),
		nodenames: cmd.StringSlice(flagNode),
	}
	return o.run(ctx)
}
