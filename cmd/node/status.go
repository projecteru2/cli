package node

import (
	"context"
	"fmt"
	"time"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
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

	logger := log.WithFunc("node.setNodeStatusOptions.run")
	logger.Error(ctx, o.heartbeat(ctx), "heartbeat")
	ticker := time.NewTicker(time.Duration(o.interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			logger.Error(ctx, o.heartbeat(ctx), "heartbeat")
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

func cmdNodeSetStatus(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	if cmd.Int("interval") < 0 {
		return fmt.Errorf("--interval must not be negative, got %d", cmd.Int("interval"))
	}

	o := &setNodeStatusOptions{
		client:   client,
		name:     cmd.StringArgs(argNode)[0],
		ttl:      cmd.Int("ttl"),
		interval: cmd.Int("interval"),
	}
	return o.run(ctx)
}
