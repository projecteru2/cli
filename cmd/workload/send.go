package workload

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
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
	logger := log.WithFunc("workload.sendWorkloadsOptions.run")
	opts := &corepb.SendOptions{
		IDs:    o.ids,
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
			logger.Errorf(ctx, errors.New(msg.Error), "send %s to %s failed", msg.Path, msg.Id)
		} else {
			logger.Infof(ctx, "send %s to %s success", msg.Path, msg.Id)
		}
	}
	return nil
}

func cmdWorkloadSend(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	content, modes, owners := utils.GenerateFileOptions(cmd)
	if len(content) == 0 {
		return fmt.Errorf("files should not be empty")
	}

	ids := cmd.Args().Slice()
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
	return o.run(ctx)
}
