package node

import (
	"context"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/pkg/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type nodeResourceOptions struct {
	client corepb.CoreRPCClient
	name   string
	fix    bool
}

func (o *nodeResourceOptions) run(ctx context.Context) error {
	resp, err := o.client.GetNodeResource(ctx, &corepb.GetNodeResourceOptions{
		Opts: &corepb.GetNodeOptions{Nodename: o.name},
		Fix:  o.fix,
	},
	)
	if err != nil {
		return err
	}

	logrus.Infof("[NodeResource] Node %s", resp.Name)
	logrus.Infof("[NodeResource] Cpu %.2f%% Memory %.2f%% Storage %.2f%% Volume %.2f%%", resp.CpuPercent*100, resp.MemoryPercent*100, resp.StoragePercent*100, resp.VolumePercent*100)
	if len(resp.Diffs) > 0 {
		for _, diff := range resp.Diffs {
			logrus.Warnf("[NodeResource] Resource diff %s", diff)
		}
	}
	return nil
}

func cmdNodeResource(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Node name must be given")
	}

	o := &nodeResourceOptions{
		client: client,
		name:   name,
		fix:    c.Bool("fix"),
	}
	return o.run(c.Context)
}
