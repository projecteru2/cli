package workload

import (
	"context"
	"errors"
	"fmt"
	"io"

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
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		if msg.Error != "" {
			return fmt.Errorf("get log of %s: %s", coreutils.ShortID(msg.Id), msg.Error)
		}

		fmt.Println(string(msg.Data))
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
		return errors.New("workload id must be specified")
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
