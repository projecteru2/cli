package workload

import (
	"context"
	"fmt"
	"io"

	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type removeWorkloadsOptions struct {
	client corepb.CoreRPCClient
	ids    []string
	step   int32
	force  bool
}

func (o *removeWorkloadsOptions) run(ctx context.Context) error {
	opts := &corepb.RemoveWorkloadOptions{
		Ids:   o.ids,
		Force: o.force,
		Step:  o.step,
	}
	resp, err := o.client.RemoveWorkload(ctx, opts)
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

		if msg.Success {
			logrus.Infof("[RemoveWorkload] %s Success", msg.Id)
		} else {
			logrus.Errorf("[RemoveWorkload] %s Failed", msg.Id)
		}
		if msg.Hook != "" {
			logrus.Info(msg.Hook)
		}
	}
	return nil
}

func cmdWorkloadRemove(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	ids := c.Args().Slice()
	if len(ids) == 0 {
		return fmt.Errorf("Workload ID(s) should not be empty")
	}

	force := c.Bool("force")
	if force {
		logrus.Warn("[RemoveWorkload] If workload not stopped, force to remove will not trigger hook process if set")
	}
	o := &removeWorkloadsOptions{
		client: client,
		ids:    ids,
		force:  force,
		step:   int32(c.Int("step")),
	}
	return o.run(c.Context)
}
