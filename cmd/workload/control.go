package workload

import (
	"context"
	"errors"
	"fmt"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	coreutils "github.com/projecteru2/core/utils"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type controlWorkloadsOptions struct {
	client corepb.CoreRPCClient
	ids    []string
	action string
	force  bool
}

func (o *controlWorkloadsOptions) run(ctx context.Context) error {
	logger := log.WithFunc("workload.controlWorkloadsOptions.run")
	opts := &corepb.ControlWorkloadOptions{
		IDs:   o.ids,
		Type:  o.action,
		Force: o.force,
	}
	resp, err := o.client.ControlWorkload(ctx, opts)
	if err != nil {
		return err
	}
	return utils.EachMessage(resp.Recv, func(msg *corepb.ControlWorkloadMessage) error {
		if msg.Hook != nil {
			logger.Infof(ctx, "hook output %s", string(msg.Hook))
		}
		if msg.Error != "" {
			return fmt.Errorf("%s %s: %s", o.action, coreutils.ShortID(msg.Id), msg.Error)
		}
		logger.Infof(ctx, "%s %s", o.action, coreutils.ShortID(msg.Id))
		return nil
	})
}

func cmdWorkloadControl(action string) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		client, err := utils.NewCoreRPCClient(ctx, cmd)
		if err != nil {
			return err
		}

		ids := cmd.Args().Slice()
		if len(ids) == 0 {
			return errors.New("workload id(s) should not be empty")
		}

		o := &controlWorkloadsOptions{
			client: client,
			ids:    ids,
			action: action,
			force:  cmd.Bool(flagForce),
		}
		return o.run(ctx)
	}
}
