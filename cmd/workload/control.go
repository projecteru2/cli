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
	var errs error
	for {
		msg, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		if msg.Hook != nil {
			logger.Infof(ctx, "hook output %s", string(msg.Hook))
		}
		if msg.Error != "" {
			errs = errors.Join(errs, fmt.Errorf("%s %s: %s", o.action, coreutils.ShortID(msg.Id), msg.Error))
			continue
		}
		logger.Infof(ctx, "%s %s", o.action, coreutils.ShortID(msg.Id))
	}
	return errs
}

func newControlWorkloadsOptions(ctx context.Context, cmd *cli.Command, action string) (*controlWorkloadsOptions, error) {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return nil, err
	}

	ids := cmd.Args().Slice()
	if len(ids) == 0 {
		return nil, errors.New("workload id(s) should not be empty")
	}

	return &controlWorkloadsOptions{
		client: client,
		ids:    ids,
		action: action,
		force:  cmd.Bool(flagForce),
	}, nil
}

func cmdWorkloadStart(ctx context.Context, cmd *cli.Command) error {
	o, err := newControlWorkloadsOptions(ctx, cmd, corecluster.WorkloadStart)
	if err != nil {
		return err
	}
	return o.run(ctx)
}

func cmdWorkloadStop(ctx context.Context, cmd *cli.Command) error {
	o, err := newControlWorkloadsOptions(ctx, cmd, corecluster.WorkloadStop)
	if err != nil {
		return err
	}
	return o.run(ctx)
}

func cmdWorkloadRestart(ctx context.Context, cmd *cli.Command) error {
	o, err := newControlWorkloadsOptions(ctx, cmd, corecluster.WorkloadRestart)
	if err != nil {
		return err
	}
	return o.run(ctx)
}
