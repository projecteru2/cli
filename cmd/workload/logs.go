package workload

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	coreutils "github.com/projecteru2/core/utils"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type workloadLogsOptions struct {
	client corepb.CoreRPCClient
	id     string
	tail   string
	since  string
	until  string
	follow bool
}

func (o *workloadLogsOptions) run(ctx context.Context) error {
	logger := log.WithFunc("workload.workloadLogsOptions.run")
	opts := &corepb.LogStreamOptions{
		Id:     o.id,
		Tail:   o.tail,
		Since:  o.since,
		Until:  o.until,
		Follow: o.follow,
	}
	resp, err := o.client.LogStream(ctx, opts)
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
			logger.Errorf(ctx, errors.New(msg.Error), "get log of %s failed", coreutils.ShortID(msg.Id))
			continue
		}

		logger.Info(ctx, string(msg.Data))
	}
	return nil
}

func cmdWorkloadLogs(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	id := cmd.Args().First()
	if id == "" {
		return fmt.Errorf("Workload ID must be specified")
	}

	o := &workloadLogsOptions{
		client: client,
		id:     id,
		tail:   cmd.String("tail"),
		since:  cmd.String("since"),
		until:  cmd.String("until"),
		follow: cmd.Bool("follow"),
	}
	return o.run(ctx)
}
