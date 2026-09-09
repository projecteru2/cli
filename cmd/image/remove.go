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

type removeImageOptions struct {
	client corepb.CoreRPCClient
	opts   *corepb.RemoveImageOptions
}

func (o *removeImageOptions) run(ctx context.Context) error {
	logger := log.WithFunc("image.removeImageOptions.run")
	resp, err := o.client.RemoveImage(ctx, o.opts)
	if err != nil {
		return err
	}

	return utils.EachMessage(resp.Recv, func(msg *corepb.RemoveImageMessage) error {
		for _, m := range msg.Messages {
			logger.Info(ctx, m)
		}
		if !msg.Success {
			return fmt.Errorf("remove %s failed", msg.Image)
		}
		logger.Infof(ctx, "remove %s success", msg.Image)
		return nil
	})
}

func cmdImageRemove(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	images := cmd.Args().Slice()
	if len(images) == 0 {
		return errors.New("images must be specified")
	}

	o := &removeImageOptions{
		client: client,
		opts: &corepb.RemoveImageOptions{
			Images:    images,
			Podname:   cmd.String(flagPod),
			Nodenames: cmd.StringSlice(flagNode),
			Prune:     cmd.Bool("prune"),
		},
	}
	return o.run(ctx)
}
