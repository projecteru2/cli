package image

import (
	"context"
	"io"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
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
	opts := &corepb.RemoveImageOptions{
		Images:    o.images,
		Step:      o.step,
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
			logrus.Infof("[CleanImage] Success remove %s", msg.Image)
		} else {
			logrus.Errorf("[Cleanimage] Failed remove %s", msg.Image)
		}
		for _, m := range msg.Messages {
			logrus.Infof(m)
		}
	}

	return nil
}

func cmdImageClean(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	images := c.Args().Slice()
	if len(images) == 0 {
		return errors.New("Images must be specified")
	}

	o := &cleanImageOptions{
		client:    client,
		images:    images,
		step:      int32(c.Int("concurrent")),
		podname:   c.String("podname"),
		nodenames: c.StringSlice("nodename"),
		prune:     c.Bool("prune"),
	}
	return o.run(c.Context)
}
