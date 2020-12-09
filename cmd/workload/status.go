package workload

import (
	"context"
	"fmt"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
)

type getWorkloadsStatusOptions struct {
	client corepb.CoreRPCClient
	ids    []string
}

func (o *getWorkloadsStatusOptions) run(ctx context.Context) error {
	resp, err := o.client.GetWorkloadsStatus(ctx, &corepb.WorkloadIDs{Ids: o.ids})
	if err != nil {
		return err
	}

	describe.WorkloadStatuses(resp.Status...)
	return nil
}

func cmdWorkloadGetStatus(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	ids := c.Args().Slice()
	if len(ids) == 0 {
		return fmt.Errorf("Workload ID(s) should not be empty")
	}

	o := &getWorkloadsStatusOptions{
		client: client,
		ids:    ids,
	}
	return o.run(c.Context)
}

type setWorkloadsStatusOptions struct {
	client corepb.CoreRPCClient
	ids    []string

	running   bool
	healthy   bool
	ttl       int64
	networks  map[string]string
	extension []byte
}

func (o *setWorkloadsStatusOptions) run(ctx context.Context) error {
	opts := &corepb.SetWorkloadsStatusOptions{Status: []*corepb.WorkloadStatus{}}
	for _, id := range o.ids {
		s := &corepb.WorkloadStatus{
			Id:        id,
			Running:   o.running,
			Healthy:   o.healthy,
			Ttl:       o.ttl,
			Networks:  o.networks,
			Extension: o.extension,
		}
		opts.Status = append(opts.Status, s)
	}

	resp, err := o.client.SetWorkloadsStatus(ctx, opts)
	if err != nil {
		return err
	}

	describe.WorkloadStatuses(resp.Status...)
	return nil
}

func cmdWorkloadSetStatus(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	ids := c.Args().Slice()
	if len(ids) == 0 {
		return fmt.Errorf("Workload ID(s) should not be empty")
	}

	o := &setWorkloadsStatusOptions{
		client:    client,
		ids:       ids,
		running:   c.Bool("running"),
		healthy:   c.Bool("healthy"),
		ttl:       c.Int64("ttl"),
		networks:  utils.SplitEquality(c.StringSlice("network")),
		extension: []byte(c.String("extension")),
	}
	return o.run(c.Context)
}
