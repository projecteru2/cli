package lambda

import (
	"context"
	"errors"
	"fmt"
	"io"

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
	printWorkloadID bool
}

func (o *runLambdaOptions) run(ctx context.Context) error {
	code, err := o.lambda(ctx)
	if err == nil {
		return cli.Exit("", code)
	}
	return err
}

func (o *runLambdaOptions) lambda(ctx context.Context) (int, error) {
	resp, err := o.client.RunAndWait(ctx)
	if err != nil {
		return -1, err
	}

	if err := resp.Send(o.opts); err != nil {
		if _, recvErr := resp.Recv(); recvErr != nil && !errors.Is(recvErr, io.EOF) {
			err = recvErr
		}
		return -1, err
	}

	iStream := interactive.NewStream(func(data []byte) error {
		return resp.Send(&corepb.RunAndWaitOptions{Cmd: data})
	}, resp.Recv, resp.CloseSend)

	go func() {
		_ = iStream.Send(newline)
	}()

	exitCount, stdin := int(o.opts.GetDeployOptions().GetCount()), o.stdin
	if o.opts.Async {
		exitCount, stdin = 0, false
	}
	return interactive.HandleStream(ctx, stdin, iStream, exitCount, o.printWorkloadID)
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
		printWorkloadID: cmd.Bool("workload-id"),
	}
	return o.run(ctx)
}

func generateLambdaOptions(cmd *cli.Command) (*corepb.RunAndWaitOptions, error) {
	if cmd.NArg() <= 0 {
		return nil, errors.New("no commands given")
	}

	network := cmd.String("network")

	memoryRequest, err := utils.ParseRAMInHuman(cmd.String("memory-request"))
	if err != nil {
		return nil, fmt.Errorf("parse memory-request: %w", err)
	}
	memoryLimit, err := utils.ParseRAMInHuman(cmd.String("memory"))
	if err != nil {
		return nil, fmt.Errorf("parse memory: %w", err)
	}

	files, err := utils.GenerateFileOptions(cmd)
	if err != nil {
		return nil, err
	}

	deployStrategy, err := utils.ParseDeployStrategy(cmd.String("deploy-strategy"))
	if err != nil {
		return nil, err
	}

	cpumem := resourcetypes.RawParams{
		"cpu-request":    cmd.Float64("cpu-request"),
		"cpu-limit":      cmd.Float64("cpu"),
		"memory-request": memoryRequest,
		"memory-limit":   memoryLimit,
	}
	storageRequest, err := utils.ParseRAMInHuman(cmd.String("storage-request"))
	if err != nil {
		return nil, fmt.Errorf("parse storage-request: %w", err)
	}
	storageLimit, err := utils.ParseRAMInHuman(cmd.String("storage"))
	if err != nil {
		return nil, fmt.Errorf("parse storage: %w", err)
	}

	resources, err := utils.EncodeResources(cmd, resourcetypes.Resources{
		utils.ResourceCPUMem:  cpumem,
		utils.ResourceStorage: utils.StorageParams(storageRequest, storageLimit, cmd.StringSlice("volume-request"), cmd.StringSlice("volume")),
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
			DeployStrategy: deployStrategy,
			Data:           files.Data,
			Owners:         files.Owners,
			Modes:          files.Modes,
			User:           cmd.String("user"),
		},
	}, nil
}
