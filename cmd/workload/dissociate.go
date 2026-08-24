package workload

import (
	"context"
	"errors"
	"io"

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
	ids := o.ids
	for _, node := range o.nodes {
		wrks, err := o.client.ListNodeWorkloads(ctx, &corepb.GetNodeOptions{Nodename: node})
		if err != nil {
			return err
		}
		for _, wrk := range wrks.Workloads {
			ids = append(ids, wrk.Id)
		}
	}
	if len(ids) == 0 {
		return errors.New("no workloads found")
	}
	opts := &corepb.DissociateWorkloadOptions{IDs: ids}
	resp, err := o.client.DissociateWorkload(ctx, opts)
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

		if msg.Error == "" {
			logger.Infof(ctx, "dissociate workload %s from eru success", msg.Id)
		} else {
			logger.Errorf(ctx, errors.New(msg.Error), "dissociate workload %s from eru failed", msg.Id)
		}
	}
	return nil
}

func cmdWorkloadDissociate(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	nodes := cmd.StringSlice("node")
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
