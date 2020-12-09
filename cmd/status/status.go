package status

import (
	"context"
	"io"
	"syscall"

	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	coreutils "github.com/projecteru2/core/utils"
	"github.com/sethvargo/go-signalcontext"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type statusOptions struct {
	client corepb.CoreRPCClient
	name   string
	entry  string
	node   string
	labels map[string]string
}

func (o *statusOptions) run(ctx context.Context) error {
	sigCtx, cancel := signalcontext.Wrap(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	resp, err := o.client.WorkloadStatusStream(sigCtx, &corepb.WorkloadStatusStreamOptions{
		Appname:    o.name,
		Entrypoint: o.entry,
		Nodename:   o.node,
		Labels:     o.labels,
	})
	if err != nil || resp == nil {
		return cli.Exit("", -1)
	}

	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil || msg == nil {
			return cli.Exit("", -1)
		}

		if msg.Error != "" {
			if msg.Delete {
				logrus.Warnf("%s deleted", coreutils.ShortID(msg.Id))
			} else {
				logrus.Errorf("[%s] status changed with error %v", coreutils.ShortID(msg.Id), msg.Error)
			}
			continue
		}

		if msg.Delete {
			logrus.Warnf("[%s] %s status expired", coreutils.ShortID(msg.Id), msg.Workload.Name)
		}

		switch {
		case !msg.Status.Running:
			logrus.Warnf("[%s] %s on %s is stopped", coreutils.ShortID(msg.Id), msg.Workload.Name, msg.Workload.Nodename)
		case !msg.Status.Healthy:
			logrus.Warnf("[%s] %s on %s is unhealthy", coreutils.ShortID(msg.Id), msg.Workload.Name, msg.Workload.Nodename)
		case msg.Status.Running && msg.Status.Healthy:
			logrus.Infof("[%s] %s back to life", coreutils.ShortID(msg.Workload.Id), msg.Workload.Name)
			for networkName, addrs := range msg.Workload.Publish {
				logrus.Infof("[%s] published at %s bind %v", coreutils.ShortID(msg.Id), networkName, addrs)
			}
		}
	}
	return nil
}

func cmdStatus(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	o := &statusOptions{
		client: client,
		name:   c.Args().First(),
		entry:  c.String("entry"),
		node:   c.String("node"),
		labels: utils.SplitEquality(c.StringSlice("label")),
	}
	return o.run(c.Context)
}
