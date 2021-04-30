package workload

import (
	"context"
	"fmt"
	"io"

	"github.com/projecteru2/cli/cmd/utils"
	corecluster "github.com/projecteru2/core/cluster"
	corepb "github.com/projecteru2/core/rpc/gen"
	coreutils "github.com/projecteru2/core/utils"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type controlWorkloadsOptions struct {
	client corepb.CoreRPCClient
	ids    []string
	action string
	force  bool
}

func (o *controlWorkloadsOptions) run(ctx context.Context) error {
	opts := &corepb.ControlWorkloadOptions{
		Ids:   o.ids,
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

		logrus.Infof("[ControlWorkload] %s", coreutils.ShortID(msg.Id))
		if msg.Hook != nil {
			logrus.Infof("[ControlWorkload] HookOutput %s", string(msg.Hook))
		}
		if msg.Error != "" {
			logrus.Errorf("[ControlWorkload] Failed %s", msg.Error)
		}
	}
	return nil
}

func createControlWorkloadsOptions(c *cli.Context, action string) (*controlWorkloadsOptions, error) {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return nil, err
	}

	ids := c.Args().Slice()
	if len(ids) == 0 {
		return nil, fmt.Errorf("Workload ID(s) should not be empty")
	}

	return &controlWorkloadsOptions{
		client: client,
		ids:    ids,
		action: action,
		force:  c.Bool("force"),
	}, nil
}

func cmdWorkloadStart(c *cli.Context) error {
	o, err := createControlWorkloadsOptions(c, corecluster.WorkloadStart)
	if err != nil {
		return err
	}
	return o.run(c.Context)
}

func cmdWorkloadStop(c *cli.Context) error {
	o, err := createControlWorkloadsOptions(c, corecluster.WorkloadStop)
	if err != nil {
		return err
	}
	return o.run(c.Context)
}

func cmdWorkloadRestart(c *cli.Context) error {
	o, err := createControlWorkloadsOptions(c, corecluster.WorkloadRestart)
	if err != nil {
		return err
	}
	return o.run(c.Context)
}
