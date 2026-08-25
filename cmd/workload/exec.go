package workload

import (
	"context"
	"errors"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/interactive"
)

type execWorkloadOptions struct {
	client      corepb.CoreRPCClient
	id          string
	interactive bool
	commands    []string
	envs        []string
	workdir     string
}

func (o *execWorkloadOptions) run(ctx context.Context) error {
	opts := &corepb.ExecuteWorkloadOptions{
		WorkloadId: o.id,
		OpenStdin:  o.interactive,
		Commands:   o.commands,
		Envs:       o.envs,
		Workdir:    o.workdir,
	}
	resp, err := o.client.ExecuteWorkload(ctx)
	if err != nil {
		return err
	}

	if err = resp.Send(opts); err != nil {
		return err
	}

	iStream := interactive.NewStream(func(data []byte) error {
		return resp.Send(&corepb.ExecuteWorkloadOptions{ReplCmd: data})
	}, resp.Recv, resp.CloseSend)

	code, err := interactive.HandleStream(ctx, opts.OpenStdin, iStream, 1, false)

	if err == nil {
		return cli.Exit("", code)
	}
	return err
}

func cmdWorkloadExec(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	id := cmd.Args().First()
	if id == "" {
		return errors.New("workload id should not be empty")
	}

	commands := cmd.Args().Tail()
	if len(commands) == 0 {
		return errors.New("commands should not be empty")
	}

	o := &execWorkloadOptions{
		client:      client,
		id:          id,
		interactive: cmd.Bool("interactive"),
		commands:    commands,
		envs:        cmd.StringSlice(flagEnv),
		workdir:     cmd.String("workdir"),
	}
	return o.run(ctx)
}
