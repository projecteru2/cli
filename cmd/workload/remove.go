package workload

import (
	"context"
	"fmt"

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

	return utils.EachMessage(resp.Recv, func(msg *corepb.RemoveWorkloadMessage) error {
		if msg.Hook != "" {
			logger.Info(ctx, msg.Hook)
		}
		if !msg.Success {
			return fmt.Errorf("remove %s failed", msg.Id)
		}
		logger.Infof(ctx, "remove %s success", msg.Id)
		return nil
	})
}

func cmdWorkloadRemove(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	ids, err := argIDs(cmd)
	if err != nil {
		return err
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
