package workload

import (
	"context"
	"errors"
	"fmt"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type sendWorkloadsOptions struct {
	client  corepb.CoreRPCClient
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

	return utils.EachMessage(resp.Recv, func(msg *corepb.SendMessage) error {
		if msg.Error != "" {
			return fmt.Errorf("send %s to %s: %s", msg.Path, msg.Id, msg.Error)
		}
		logger.Infof(ctx, "send %s to %s success", msg.Path, msg.Id)
		return nil
	})
}

func cmdWorkloadSend(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	files, err := utils.GenerateFileOptions(cmd)
	if err != nil {
		return err
	}
	if len(files.Data) == 0 {
		return errors.New("files should not be empty")
	}

	ids := cmd.Args().Slice()
	if len(ids) == 0 {
		return errors.New("workload id(s) should not be empty")
	}

	o := &sendWorkloadsOptions{
		client:  client,
		ids:     ids,
		content: files.Data,
		modes:   files.Modes,
		owners:  files.Owners,
	}
	return o.run(ctx)
}
