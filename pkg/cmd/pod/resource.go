package pod

import (
	"context"
	"strings"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/pkg/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type resourcePodOptions struct {
	client corepb.CoreRPCClient
	name   string
}

func (o *resourcePodOptions) run(ctx context.Context) error {
	resp, err := o.client.GetPodResource(ctx, &corepb.GetPodOptions{
		Name: o.name,
	})
	if err != nil {
		return err
	}

	logrus.Infof("[PodResource] Pod %s", resp.Name)
	for _, nodeResource := range resp.NodesResource {
		logrus.Infof("[PodResource] Node %s Cpu %.2f%% Memory %.2f%% Storage %.2f%% Volume %.2f%%",
			nodeResource.Name, nodeResource.CpuPercent*100, nodeResource.MemoryPercent*100,
			nodeResource.StoragePercent*100, nodeResource.VolumePercent*100,
		)
		if len(nodeResource.Diffs) > 0 {
			logrus.Warnf("[PodResource] Node %s resource diff %s", nodeResource.Name, strings.Join(nodeResource.Diffs, "\n"))
		}
	}
	return nil
}

func cmdPodResource(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Pod name must be given")
	}

	o := &resourcePodOptions{
		client: client,
		name:   name,
	}
	return o.run(c.Context)
}
