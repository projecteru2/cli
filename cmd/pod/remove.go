package pod

import (
	"context"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type removePodOptions struct {
	client corepb.CoreRPCClient
	name   string
}

func (o *removePodOptions) run(ctx context.Context) error {
	_, err := o.client.RemovePod(ctx, &corepb.RemovePodOptions{
		Name: o.name,
	})
	if err != nil {
		return err
	}

	logrus.Infof("[RemovePod] success, name: %s", o.name)
	return nil
}

func cmdPodRemove(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Pod name must be given")
	}

	o := &removePodOptions{
		client: client,
		name:   name,
	}
	return o.run(c.Context)
}
