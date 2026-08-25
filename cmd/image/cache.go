package image

import (
	"context"
	"errors"
	"fmt"

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

	return utils.EachMessage(resp.Recv, func(msg *corepb.CacheImageMessage) error {
		if !msg.Success {
			return fmt.Errorf("cache image %s on %s: %s", msg.Image, msg.Nodename, msg.Message)
		}
		logger.Infof(ctx, "cache image %s on %s success", msg.Image, msg.Nodename)
		return nil
	})
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
