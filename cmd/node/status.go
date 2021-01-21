package node

import (
	"context"
	"time"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
)

type setNodeStatusOptions struct {
	client   corepb.CoreRPCClient
	name     string
	ttl      int
	interval int
}

func (o *setNodeStatusOptions) run(ctx context.Context) error {
	if o.interval == 0 {
		return o.heartbeat(ctx)
	}

	timer := time.NewTicker(time.Duration(o.interval) * time.Second)
	defer timer.Stop()

	var err error
	for {
		select {
		case <-ctx.Done():
			return err
		case <-timer.C:
			err = o.heartbeat(ctx)
		}
	}
}

func (o *setNodeStatusOptions) heartbeat(ctx context.Context) error {
	_, err := o.client.SetNodeStatus(ctx, &corepb.SetNodeStatusOptions{
		Nodename: o.name,
		Ttl:      int64(o.ttl),
	})
	return err
}

func cmdNodeSetStatus(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Node name must be given")
	}

	o := &setNodeStatusOptions{
		client:   client,
		name:     name,
		ttl:      c.Int("ttl"),
		interval: c.Int("interval"),
	}
	return o.run(c.Context)
}
