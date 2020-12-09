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

type cacheImageOptions struct {
	client    corepb.CoreRPCClient
	images    []string
	step      int32
	podname   string
	nodenames []string
}

func (o *cacheImageOptions) run(ctx context.Context) error {
	opts := &corepb.CacheImageOptions{
		Images:    o.images,
		Step:      o.step,
		Podname:   o.podname,
		Nodenames: o.nodenames,
	}
	resp, err := o.client.CacheImage(ctx, opts)
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
			logrus.Infof("[CacheImage] cache image %s on %s success", msg.Image, msg.Nodename)
		} else {
			logrus.Warnf("[CacheImage] cache image %s on %s failed", msg.Image, msg.Nodename)
		}
	}
	return nil
}

func cmdImageCache(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	images := c.Args().Slice()
	if len(images) == 0 {
		return errors.New("Images must be specified")
	}

	o := &cacheImageOptions{
		client:    client,
		images:    images,
		step:      int32(c.Int("concurrent")),
		podname:   c.String("podname"),
		nodenames: c.StringSlice("nodename"),
	}
	return o.run(c.Context)
}
