package pod

import (
	"context"
	"errors"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
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

func cmdPodAdd(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	name := cmd.Args().First()
	if name == "" {
		return errors.New("Pod name must be given")
	}

	o := &addPodOptions{
		client: client,
		name:   name,
		desc:   cmd.String("desc"),
	}
	return o.run(ctx)
}
