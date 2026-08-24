package lambda

import (
	"context"
	"errors"
	"fmt"
	"strings"

	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/interactive"
)

var newline = []byte{'\n'}

type runLambdaOptions struct {
	client          corepb.CoreRPCClient
	opts            *corepb.RunAndWaitOptions
	stdin           bool
	count           int
	printWorkloadID bool
}

func (o *runLambdaOptions) run(ctx context.Context) error {
	code, err := lambda(ctx, o.client, o.opts, o.stdin, o.count, o.printWorkloadID)
	if err == nil {
		return cli.Exit("", code)
	}
	return err
}

func cmdLambdaRun(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	opts, err := generateLambdaOptions(cmd)
	if err != nil {
		return err
	}

	o := &runLambdaOptions{
		client:          client,
		opts:            opts,
		stdin:           cmd.Bool("stdin"),
		count:           cmd.Int("count"),
		printWorkloadID: cmd.Bool("workload-id"),
	}
	return o.run(ctx)
}

func lambda(ctx context.Context, client corepb.CoreRPCClient, opts *corepb.RunAndWaitOptions, stdin bool, count int, printWorkloadID bool) (code int, err error) {
	resp, err := client.RunAndWait(ctx)
	if err != nil {
		return -1, err
	}

	if err := resp.Send(opts); err != nil {
		return -1, err
	}

	iStream := interactive.Stream{
		Recv: resp.Recv,
		Send: func(data []byte) error {
			return resp.Send(&corepb.RunAndWaitOptions{Cmd: data})
		},
	}

	go func() {
		_ = iStream.Send(newline)
	}()

	return interactive.HandleStream(ctx, stdin, iStream, count, printWorkloadID)
}

func generateLambdaOptions(cmd *cli.Command) (*corepb.RunAndWaitOptions, error) {
	if cmd.NArg() <= 0 {
		return nil, errors.New("no commands given")
	}

	network := cmd.String("network")

	memoryRequest, err := utils.ParseRAMInHuman(cmd.String("memory-request"))
	if err != nil {
		return nil, fmt.Errorf("parse memory: %w", err)
	}
	memoryLimit, err := utils.ParseRAMInHuman(cmd.String("memory"))
	if err != nil {
		return nil, fmt.Errorf("parse memory: %w", err)
	}

	content, modes, owners := utils.GenerateFileOptions(cmd)

	cpumem := resourcetypes.RawParams{
		"cpu-request":    cmd.Float64("cpu-request"),
		"cpu-limit":      cmd.Float64("cpu"),
		"memory-request": memoryRequest,
		"memory-limit":   memoryLimit,
	}
	storageRequest, err := utils.ParseRAMInHuman(cmd.String("storage-request"))
	if err != nil {
		return nil, fmt.Errorf("parse storage: %w", err)
	}
	storageLimit, err := utils.ParseRAMInHuman(cmd.String("storage"))
	if err != nil {
		return nil, fmt.Errorf("parse storage: %w", err)
	}

	storage := resourcetypes.RawParams{
		"storage-request": storageRequest,
		"storage-limit":   storageLimit,
		"volumes-request": cmd.StringSlice("volume-request"),
		"volumes-limit":   cmd.StringSlice("volume"),
	}

	resources, err := utils.EncodeResources(cmd, resourcetypes.Resources{
		resourceCPUMem:  cpumem,
		resourceStorage: storage,
	})
	if err != nil {
		return nil, err
	}

	return &corepb.RunAndWaitOptions{
		Async:        cmd.Bool("async"),
		AsyncTimeout: int32(cmd.Int("async-timeout")), //nolint:gosec
		DeployOptions: &corepb.DeployOptions{
			Name: "lambda",
			Entrypoint: &corepb.EntrypointOptions{
				Name:       cmd.String("name"),
				Commands:   cmd.Args().Slice(),
				Privileged: cmd.Bool("privileged"),
				Dir:        cmd.String("working-dir"),
			},
			Resources: resources,
			Podname:   cmd.String("pod"),
			NodeFilter: &corepb.NodeFilter{
				Includes: cmd.StringSlice("node"),
			},
			Image:          cmd.String("image"),
			Count:          int32(cmd.Int("count")), //nolint:gosec
			Env:            cmd.StringSlice("env"),
			Networks:       utils.GetNetworks(network),
			OpenStdin:      cmd.Bool("stdin"),
			DeployStrategy: corepb.DeployOptions_Strategy(corepb.DeployOptions_Strategy_value[strings.ToUpper(cmd.String("deploy-strategy"))]),
			Data:           content,
			Owners:         owners,
			Modes:          modes,
			User:           cmd.String("user"),
		},
	}, nil
}
