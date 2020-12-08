package workload

import (
	"context"

	"github.com/projecteru2/cli/pkg/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type reallocWorkloadsOptions struct {
	client corepb.CoreRPCClient
	opts   *corepb.ReallocOptions
}

func (o *reallocWorkloadsOptions) run(ctx context.Context) error {
	resp, err := o.client.ReallocResource(ctx, o.opts)
	if err != nil {
		return err
	}

	if resp.Error != "" {
		logrus.Infof("[Realloc] Failed by %+v", resp.Error)
	} else {
		logrus.Info("[Realloc] Success")
	}
	return nil
}

func cmdWorkloadRealloc(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	opts, err := utils.GenerateReallocOptions(c)
	if err != nil {
		return err
	}

	o := &reallocWorkloadsOptions{
		client: client,
		opts:   opts,
	}
	return o.run(c.Context)
}
