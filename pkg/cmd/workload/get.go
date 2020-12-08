package workload

import (
	"context"
	"fmt"

	"github.com/projecteru2/cli/pkg/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type getWorkloadsOptions struct {
	client corepb.CoreRPCClient
	ids    []string
}

func (o *getWorkloadsOptions) run(ctx context.Context) error {
	resp, err := o.client.GetWorkloads(ctx, &corepb.WorkloadIDs{Ids: o.ids})
	if err != nil {
		return err
	}

	logrus.Info(resp.Workloads)
	return nil
}

func cmdWorkloadGet(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	ids := c.Args().Slice()
	if len(ids) == 0 {
		return fmt.Errorf("Workload ID(s) should not be empty")
	}

	o := &getWorkloadsOptions{
		client: client,
		ids:    ids,
	}
	return o.run(c.Context)
}
