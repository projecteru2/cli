package workload

import (
	"context"
	"errors"
	"io"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type removeWorkloadsOptions struct {
	client corepb.CoreRPCClient
	ids    []string
	force  bool
}

func (o *removeWorkloadsOptions) run(ctx context.Context) error {
	logger := log.WithFunc("workload.removeWorkloadsOptions.run")
	opts := &corepb.RemoveWorkloadOptions{
		IDs:   o.ids,
		Force: o.force,
	}
	resp, err := o.client.RemoveWorkload(ctx, opts)
	if err != nil {
		return err
	}

	for {
		msg, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		if msg.Success {
			logger.Infof(ctx, "remove %s success", msg.Id)
		} else {
			logger.Warnf(ctx, "remove %s failed", msg.Id)
		}
		if msg.Hook != "" {
			logger.Info(ctx, msg.Hook)
		}
	}
	return nil
}

func cmdWorkloadRemove(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	ids := cmd.Args().Slice()
	if len(ids) == 0 {
		return errors.New("workload id(s) should not be empty")
	}

	force := cmd.Bool(flagForce)
	if force {
		log.WithFunc("workload.cmdWorkloadRemove").Warn(ctx, "if workload not stopped, force to remove will not trigger hook process if set")
	}
	o := &removeWorkloadsOptions{
		client: client,
		ids:    ids,
		force:  force,
	}
	return o.run(ctx)
}
