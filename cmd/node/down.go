package node

import (
	"context"
	"errors"
	"time"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type setNodeDownOptions struct {
	client       corepb.CoreRPCClient
	name         string
	check        bool
	checkTimeout int
}

func (o *setNodeDownOptions) run(ctx context.Context) error {
	logger := log.WithFunc("node.setNodeDownOptions.run")
	if o.check {
		checkCtx, cancel := context.WithTimeout(ctx, time.Duration(o.checkTimeout)*time.Second)
		defer cancel()
		status, err := o.client.GetNodeStatus(checkCtx, &corepb.GetNodeStatusOptions{Nodename: o.name})
		if err == nil && status.Alive {
			return errors.New("node is still alive, not marking it down")
		}
	}

	if _, err := o.client.SetNode(ctx, &corepb.SetNodeOptions{
		Nodename: o.name,
		Bypass:   corepb.TriOpt_TRUE,
	}); err != nil {
		return err
	}
	logger.Infof(ctx, "node %s down", o.name)
	return nil
}

func cmdNodeSetDown(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	name := cmd.Args().First()
	if name == "" {
		return errors.New("node name must be given")
	}

	o := &setNodeDownOptions{
		client:       client,
		name:         name,
		check:        cmd.Bool("check"),
		checkTimeout: cmd.Int("check-timeout"),
	}
	return o.run(ctx)
}
