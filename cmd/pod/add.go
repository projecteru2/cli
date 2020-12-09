package pod

import (
	"context"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"
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

	describe.Pods(pod)
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
