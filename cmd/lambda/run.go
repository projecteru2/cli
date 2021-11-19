package lambda

import (
	"context"
	"strings"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/interactive"
	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/juju/errors"
	"github.com/urfave/cli/v2"
)

type runLambdaOptions struct {
	client          corepb.CoreRPCClient
	opts            *corepb.RunAndWaitOptions
	stdin           bool
	count           int
	printWorkloadID bool
}

func (o *runLambdaOptions) run(_ context.Context) error {
	code, err := lambda(o.client, o.opts, o.stdin, o.count, o.printWorkloadID)
	if err == nil {
		return cli.Exit("", code)
	}
	return err
}

func cmdLambdaRun(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	opts, err := generateLambdaOptions(c)
	if err != nil {
		return err
	}

	o := &runLambdaOptions{
		client:          client,
		opts:            opts,
		stdin:           c.Bool("stdin"),
		count:           c.Int("count"),
		printWorkloadID: c.Bool("workload-id"),
	}
	return o.run(c.Context)
}

var clrf = []byte{0xa}

func lambda(client corepb.CoreRPCClient, opts *corepb.RunAndWaitOptions, stdin bool, count int, printWorkloadID bool) (code int, err error) {
	resp, err := client.RunAndWait(context.Background())
	if err != nil {
		return -1, err
	}

	if resp.Send(opts) != nil {
		return -1, err
	}

	iStream := interactive.Stream{
		Recv: resp.Recv,
		Send: func(cmd []byte) error {
			return resp.Send(&corepb.RunAndWaitOptions{Cmd: cmd})
		},
	}

	go func() {
		_ = iStream.Send(clrf)
	}()

	return interactive.HandleStream(stdin, iStream, count, printWorkloadID)
}

func generateLambdaOptions(c *cli.Context) (*corepb.RunAndWaitOptions, error) {
	if c.NArg() <= 0 {
		return nil, errors.New("[Lambda] no commands")
	}

	network := c.String("network")

	content, modes, owners := utils.GenerateFileOptions(c)

	stringFlags := []string{"cpu-request", "cpu", "memory-request", "memory", "storage-request", "storage"}
	stringSliceFlags := []string{"volume-request", "volume"}
	resourceOpts := utils.GetResourceOpts(c, stringFlags, stringSliceFlags, nil, nil)

	for _, flag := range []string{"cpu", "memory", "storage", "volume"} {
		resourceOpts[flag+"-limit"] = resourceOpts[flag]
		delete(resourceOpts, flag)
	}

	return &corepb.RunAndWaitOptions{
		Async:        c.Bool("async"),
		AsyncTimeout: int32(c.Int("async-timeout")),
		DeployOptions: &corepb.DeployOptions{
			Name: "lambda",
			Entrypoint: &corepb.EntrypointOptions{
				Name:       c.String("name"),
				Commands:   c.Args().Slice(),
				Privileged: c.Bool("privileged"),
				Dir:        c.String("working-dir"),
			},
			ResourceOpts: resourceOpts,
			Podname:      c.String("pod"),
			NodeFilter: &corepb.NodeFilter{
				Includes: c.StringSlice("node"),
			},
			Image:          c.String("image"),
			Count:          int32(c.Int("count")),
			Env:            c.StringSlice("env"),
			Networks:       utils.GetNetworks(network),
			OpenStdin:      c.Bool("stdin"),
			DeployStrategy: corepb.DeployOptions_Strategy(corepb.DeployOptions_Strategy_value[strings.ToUpper(c.String("deploy-strategy"))]),
			Data:           content,
			Owners:         owners,
			Modes:          modes,
			User:           c.String("user"),
		},
	}, nil
}
