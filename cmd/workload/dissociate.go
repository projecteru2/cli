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
}

func (o *dissociateWorkloadsOptions) run(ctx context.Context) error {
	opts := &corepb.DissociateWorkloadOptions{Ids: o.ids}
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

	ids := c.Args().Slice()
	if len(ids) == 0 {
		return fmt.Errorf("Workload ID(s) should not be empty")
	}

	o := &dissociateWorkloadsOptions{
		client: client,
		ids:    ids,
	}
	return o.run(c.Context)
}
