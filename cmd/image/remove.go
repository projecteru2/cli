package image

import (
	"context"
	"errors"
	"io"

	"github.com/urfave/cli/v3"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/projecteru2/cli/cmd/utils"
)

type cleanImageOptions struct {
	client    corepb.CoreRPCClient
	images    []string
	step      int32
	podname   string
	nodenames []string
	prune     bool
}

func (o *cleanImageOptions) run(ctx context.Context) error {
	logger := log.WithFunc("image.cleanImageOptions.run")
	opts := &corepb.RemoveImageOptions{
		Images:    o.images,
		Podname:   o.podname,
		Nodenames: o.nodenames,
		Prune:     o.prune,
	}
	resp, err := o.client.RemoveImage(ctx, opts)
	if err != nil {
		return err
	}

	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if msg.Success {
			logger.Infof(ctx, "remove %s success", msg.Image)
		} else {
			logger.Errorf(ctx, errors.New("remove image failed"), "remove %s", msg.Image)
		}
		for _, m := range msg.Messages {
			logger.Info(ctx, m)
		}
	}

	return nil
}

func cmdImageClean(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	images := cmd.Args().Slice()
	if len(images) == 0 {
		return errors.New("Images must be specified")
	}

	o := &cleanImageOptions{
		client:    client,
		images:    images,
		step:      int32(cmd.Int("concurrent")),
		podname:   cmd.String("pod"),
		nodenames: cmd.StringSlice("node"),
		prune:     cmd.Bool("prune"),
	}
	return o.run(ctx)
}
