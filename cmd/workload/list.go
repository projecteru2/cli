package workload

import (
	"context"
	"io"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
)

type listWorkloadsOptions struct {
	client corepb.CoreRPCClient
	// must be set
	appname string
	limit   int64
	// filters
	entrypoint string
	nodename   string
	labels     map[string]string
}

func (o *listWorkloadsOptions) run(ctx context.Context) error {
	opts := &corepb.ListWorkloadsOptions{
		Appname:    o.appname,
		Entrypoint: o.entrypoint,
		Nodename:   o.nodename,
		Labels:     o.labels,
		Limit:      o.limit,
	}

	resp, err := o.client.ListWorkloads(ctx, opts)
	if err != nil {
		return err
	}

	workloads := []*corepb.Workload{}
	for {
		w, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		workloads = append(workloads, w)
	}

	describe.Workloads(workloads...)
	return nil
}

func cmdWorkloadList(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	o := &listWorkloadsOptions{
		client:     client,
		appname:    c.Args().First(),
		entrypoint: c.String("entry"),
		nodename:   c.String("nodename"),
		labels:     utils.SplitEquality(c.StringSlice("label")),
		limit:      c.Int64("limit"),
	}
	return o.run(c.Context)
}
