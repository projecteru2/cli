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

type sendWorkloadsOptions struct {
	client corepb.CoreRPCClient
	// workload ids
	ids     []string
	content map[string][]byte
	modes   map[string]*corepb.FileMode
	owners  map[string]*corepb.FileOwner
}

func (o *sendWorkloadsOptions) run(ctx context.Context) error {
	opts := &corepb.SendOptions{
		Ids:    o.ids,
		Data:   o.content,
		Modes:  o.modes,
		Owners: o.owners,
	}
	resp, err := o.client.Send(ctx, opts)
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

		if msg.Error != "" {
			logrus.Errorf("[Send] Failed send %s to %s", msg.Path, msg.Id)
		} else {
			logrus.Infof("[Send] Send %s to %s success", msg.Path, msg.Id)
		}
	}
	return nil
}

func cmdWorkloadSend(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	content, modes, owners := utils.GenerateFileOptions(c)
	if len(content) == 0 {
		return fmt.Errorf("files should not be empty")
	}

	ids := c.Args().Slice()
	if len(ids) == 0 {
		return fmt.Errorf("Workload ID(s) should not be empty")
	}

	o := &sendWorkloadsOptions{
		client:  client,
		ids:     ids,
		content: content,
		modes:   modes,
		owners:  owners,
	}
	return o.run(c.Context)
}
