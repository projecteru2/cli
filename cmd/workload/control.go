package workload

import (
	"context"
	"errors"
	"fmt"
	"io"

	corecluster "github.com/projecteru2/core/cluster"
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
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		logger.Infof(ctx, "%s %s", o.action, coreutils.ShortID(msg.Id))
		if msg.Hook != nil {
			logger.Infof(ctx, "hook output %s", string(msg.Hook))
		}
		if msg.Error != "" {
			logger.Error(ctx, errors.New(msg.Error))
		}
	}
	return nil
}

func createControlWorkloadsOptions(ctx context.Context, cmd *cli.Command, action string) (*controlWorkloadsOptions, error) {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return nil, err
	}

	ids := cmd.Args().Slice()
	if len(ids) == 0 {
		return nil, fmt.Errorf("Workload ID(s) should not be empty")
	}

	return &controlWorkloadsOptions{
		client: client,
		ids:    ids,
		action: action,
		force:  cmd.Bool("force"),
	}, nil
}

func cmdWorkloadStart(ctx context.Context, cmd *cli.Command) error {
	o, err := createControlWorkloadsOptions(ctx, cmd, corecluster.WorkloadStart)
	if err != nil {
		return err
	}
	return o.run(ctx)
}

func cmdWorkloadStop(ctx context.Context, cmd *cli.Command) error {
	o, err := createControlWorkloadsOptions(ctx, cmd, corecluster.WorkloadStop)
	if err != nil {
		return err
	}
	return o.run(ctx)
}

func cmdWorkloadRestart(ctx context.Context, cmd *cli.Command) error {
	o, err := createControlWorkloadsOptions(ctx, cmd, corecluster.WorkloadRestart)
	if err != nil {
		return err
	}
	return o.run(ctx)
}
