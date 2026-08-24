package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/interactive"
)

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

var clrf = []byte{0xa}

func lambda(ctx context.Context, client corepb.CoreRPCClient, opts *corepb.RunAndWaitOptions, stdin bool, count int, printWorkloadID bool) (code int, err error) {
	resp, err := client.RunAndWait(ctx)
	if err != nil {
		return -1, err
	}

	if resp.Send(opts) != nil {
		return -1, err
	}

	iStream := interactive.Stream{
		Recv: resp.Recv,
		Send: func(data []byte) error {
			return resp.Send(&corepb.RunAndWaitOptions{Cmd: data})
		},
	}

	go func() {
		_ = iStream.Send(clrf)
	}()

	return interactive.HandleStream(ctx, stdin, iStream, count, printWorkloadID)
}

func generateLambdaOptions(cmd *cli.Command) (*corepb.RunAndWaitOptions, error) {
	if cmd.NArg() <= 0 {
		return nil, errors.New("[Lambda] no commands")
	}

	network := cmd.String("network")

	memoryRequest, err := utils.ParseRAMInHuman(cmd.String("memory-request"))
	if err != nil {
		return nil, fmt.Errorf("[Lambda] memory wrong %v", err)
	}
	memoryLimit, err := utils.ParseRAMInHuman(cmd.String("memory"))
	if err != nil {
		return nil, fmt.Errorf("[Lambda] memory wrong %v", err)
	}

	content, modes, owners := utils.GenerateFileOptions(cmd)

	cpumem := resourcetypes.RawParams{
		"cpu-request":    cmd.Float64("cpu-request"),
		"cpu-limit":      cmd.Float64("cpu"),
		"memory-request": memoryRequest,
		"memory-limit":   memoryLimit,
	}
	storage := resourcetypes.RawParams{
		"storage-request": cmd.Int64("storage-request"),
		"storage-limit":   cmd.Int64("storage"),
		"volumes-request": cmd.StringSlice("volumes-request"),
		"volumes-limit":   cmd.StringSlice("volumes"),
	}

	if cmd.Bool("cpu-bind") {
		cpumem["cpu-bind"] = true
	}

	cb, _ := json.Marshal(cpumem)
	sb, _ := json.Marshal(storage)

	resources := map[string][]byte{
		"cpumem":  cb,
		"storage": sb,
	}

	return &corepb.RunAndWaitOptions{
		Async:        cmd.Bool("async"),
		AsyncTimeout: int32(cmd.Int("async-timeout")),
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
			Count:          int32(cmd.Int("count")),
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
