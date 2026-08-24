package status

import (
	"context"
	"errors"
	"io"
	"os/signal"
	"syscall"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	coreutils "github.com/projecteru2/core/utils"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type statusOptions struct {
	client corepb.CoreRPCClient
	name   string
	entry  string
	node   string
	labels map[string]string
}

func (o *statusOptions) run(ctx context.Context) error {
	logger := log.WithFunc("status.statusOptions.run")
	sigCtx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	resp, err := o.client.WorkloadStatusStream(sigCtx, &corepb.WorkloadStatusStreamOptions{
		Appname:    o.name,
		Entrypoint: o.entry,
		Nodename:   o.node,
		Labels:     o.labels,
	})
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
			if msg.Delete {
				logger.Warnf(ctx, "%s deleted", coreutils.ShortID(msg.Id))
			} else {
				logger.Errorf(ctx, errors.New(msg.Error), "[%s] status changed with error", coreutils.ShortID(msg.Id))
			}
			continue
		}

		if msg.Delete {
			logger.Warnf(ctx, "[%s] %s status expired", coreutils.ShortID(msg.Id), msg.Workload.Name)
		}

		switch {
		case !msg.Status.Running:
			logger.Warnf(ctx, "[%s] %s on %s is stopped", coreutils.ShortID(msg.Id), msg.Workload.Name, msg.Workload.Nodename)
		case !msg.Status.Healthy:
			logger.Warnf(ctx, "[%s] %s on %s is unhealthy", coreutils.ShortID(msg.Id), msg.Workload.Name, msg.Workload.Nodename)
		default:
			logger.Infof(ctx, "[%s] %s back to life", coreutils.ShortID(msg.Workload.Id), msg.Workload.Name)
			for networkName, addrs := range msg.Workload.Publish {
				logger.Infof(ctx, "[%s] published at %s bind %v", coreutils.ShortID(msg.Id), networkName, addrs)
			}
		}
	}
	return nil
}

func cmdStatus(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	o := &statusOptions{
		client: client,
		name:   cmd.Args().First(),
		entry:  cmd.String("entry"),
		node:   cmd.String("node"),
		labels: utils.SplitEquality(cmd.StringSlice("label")),
	}
	return o.run(ctx)
}
