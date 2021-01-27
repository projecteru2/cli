package workload

import (
	"context"
	"fmt"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/interactive"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
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

	if err := resp.Send(opts); err != nil {
		return err
	}

	iStream := interactive.Stream{
		Recv: resp.Recv,
		Send: func(cmd []byte) error {
			return resp.Send(&corepb.ExecuteWorkloadOptions{ReplCmd: cmd})
		},
	}

	code, err := interactive.HandleStream(opts.OpenStdin, iStream, 1, false)

	if err == nil {
		return cli.Exit("", code)
	}
	return err
}

func cmdWorkloadExec(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	id := c.Args().First()
	if id == "" {
		return fmt.Errorf("Workload ID should not be empty")
	}

	commands := c.Args().Tail()
	if len(commands) == 0 {
		return fmt.Errorf("Commands should not be empty")
	}

	o := &execWorkloadOptions{
		client:      client,
		id:          id,
		interactive: c.Bool("interactive"),
		commands:    commands,
		envs:        c.StringSlice("env"),
		workdir:     c.String("workdir"),
	}
	return o.run(c.Context)
}
