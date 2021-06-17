package workload

import (
	"context"
	"fmt"
	"io"

	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	coreutils "github.com/projecteru2/core/utils"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
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
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if msg.Error != "" {
			logrus.Errorf("[GetWorkloadLog] Failed %s %s", coreutils.ShortID(msg.Id), msg.Error)
			continue
		}

		logrus.Infof("[GetWorkloadLog] %s", string(msg.Data))
	}
	return nil
}

func cmdWorkloadLogs(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	id := c.Args().First()
	if id == "" {
		return fmt.Errorf("Workload ID must be specified")
	}

	o := &workloadLogsOptions{
		client: client,
		id:     id,
		tail:   c.String("tail"),
		since:  c.String("since"),
		until:  c.String("until"),
		follow: c.Bool("follow"),
	}
	return o.run(c.Context)
}
