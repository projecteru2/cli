package workload

import (
	"context"
	"errors"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
)

type getWorkloadsOptions struct {
	client corepb.CoreRPCClient
	ids    []string
}

func (o *getWorkloadsOptions) run(ctx context.Context) error {
	resp, err := o.client.GetWorkloads(ctx, &corepb.WorkloadIDs{IDs: o.ids})
	if err != nil {
		return err
	}

	describe.Workloads(resp.Workloads...)
	return nil
}

func cmdWorkloadGet(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	ids := cmd.Args().Slice()
	if len(ids) == 0 {
		return errors.New("workload id(s) should not be empty")
	}

	o := &getWorkloadsOptions{
		client: client,
		ids:    ids,
	}
	return o.run(ctx)
}
