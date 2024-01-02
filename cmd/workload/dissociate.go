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

type dissociateWorkloadsOptions struct {
	client corepb.CoreRPCClient
	ids    []string
	nodes  []string
}

func (o *dissociateWorkloadsOptions) run(ctx context.Context) error {
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
		return fmt.Errorf("no workloads found")
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
			logrus.Infof("[Dissociate] Dissociate workload %s from eru success", msg.Id)
		} else {
			logrus.Errorf("[Dissociate] Dissociate workload %s from eru failed %v", msg.Id, msg.Error)
		}
	}
	return nil
}

func cmdWorkloadDissociate(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	nodes := c.StringSlice("node")
	ids := c.Args().Slice()
	if len(ids) == 0 && len(nodes) == 0 {
		return fmt.Errorf("Workload ID(s) and Node(s) should not be empty")
	}

	o := &dissociateWorkloadsOptions{
		client: client,
		ids:    ids,
		nodes:  nodes,
	}
	return o.run(c.Context)
}
