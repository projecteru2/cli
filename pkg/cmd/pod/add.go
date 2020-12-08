package pod

import (
	"context"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/pkg/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type addPodOptions struct {
	client corepb.CoreRPCClient
	name   string
	desc   string
}

func (o *addPodOptions) run(ctx context.Context) error {
	pod, err := o.client.AddPod(ctx, &corepb.AddPodOptions{
		Name: o.name,
		Desc: o.desc,
	})
	if err != nil {
		return err
	}

	logrus.Infof("[AddPod] success, name: %s, desc: %s", pod.GetName(), pod.GetDesc())
	return nil
}

func cmdPodAdd(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Pod name must be given")
	}

	o := &addPodOptions{
		client: client,
		name:   name,
		desc:   c.String("desc"),
	}
	return o.run(c.Context)
}
