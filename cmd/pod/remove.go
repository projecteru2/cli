package pod

import (
	"context"
	"errors"

	"github.com/urfave/cli/v3"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/projecteru2/cli/cmd/utils"
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

	log.WithFunc("pod.removePodOptions.run").Infof(ctx, "remove pod %s success", o.name)
	return nil
}

func cmdPodRemove(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	name := cmd.Args().First()
	if name == "" {
		return errors.New("Pod name must be given")
	}

	o := &removePodOptions{
		client: client,
		name:   name,
	}
	return o.run(ctx)
}
