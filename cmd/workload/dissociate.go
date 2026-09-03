package workload

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type dissociateWorkloadsOptions struct {
	client corepb.CoreRPCClient
	ids    []string
	nodes  []string
}

func (o *dissociateWorkloadsOptions) run(ctx context.Context) error {
	logger := log.WithFunc("workload.dissociateWorkloadsOptions.run")
	ids := slices.Clone(o.ids)
	for _, node := range o.nodes {
		wrks, err := o.client.ListNodeWorkloads(ctx, &corepb.GetNodeOptions{Nodename: node})
		if err != nil {
			return err
		}
		for _, wrk := range wrks.Workloads {
			ids = append(ids, wrk.Id)
		}
	}
	ids = slices.Compact(slices.Sorted(slices.Values(ids)))
	if len(ids) == 0 {
		return errors.New("no workloads found")
	}
	opts := &corepb.DissociateWorkloadOptions{IDs: ids}
	resp, err := o.client.DissociateWorkload(ctx, opts)
	if err != nil {
		return err
	}

	return utils.EachMessage(resp.Recv, func(msg *corepb.DissociateWorkloadMessage) error {
		if msg.Error != "" {
			return fmt.Errorf("dissociate workload %s: %s", msg.Id, msg.Error)
		}
		logger.Infof(ctx, "dissociate workload %s from eru success", msg.Id)
		return nil
	})
}

func cmdWorkloadDissociate(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	nodes := cmd.StringSlice(flagNode)
	ids := cmd.Args().Slice()
	if len(ids) == 0 && len(nodes) == 0 {
		return errors.New("workload id(s) or node(s) should not be empty")
	}

	o := &dissociateWorkloadsOptions{
		client: client,
		ids:    ids,
		nodes:  nodes,
	}
	return o.run(ctx)
}
