package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/juju/errors"
	"github.com/urfave/cli/v2"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/interactive"
	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
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

	memoryRequest, err := utils.ParseRAMInHuman(c.String("memory-request"))
	if err != nil {
		return nil, fmt.Errorf("[Lambda] memory wrong %v", err)
	}
	memoryLimit, err := utils.ParseRAMInHuman(c.String("memory"))
	if err != nil {
		return nil, fmt.Errorf("[Lambda] memory wrong %v", err)
	}

	content, modes, owners := utils.GenerateFileOptions(c)

	cpumem := resourcetypes.RawParams{
		"cpu-request":    c.Float64("cpu-request"),
		"cpu-limit":      c.Float64("cpu"),
		"memory-request": memoryRequest,
		"memory-limit":   memoryLimit,
	}
	storage := resourcetypes.RawParams{
		"storage-request": c.Int64("storage-request"),
		"storage-limit":   c.Int64("storage"),
		"volumes-request": c.StringSlice("volumes-request"),
		"volumes-limit":   c.StringSlice("volumes"),
	}

	if c.Bool("cpu-bind") {
		cpumem["cpu-bind"] = true
	}

	cb, _ := json.Marshal(cpumem)
	sb, _ := json.Marshal(storage)

	resources := map[string][]byte{
		"cpumem":  cb,
		"storage": sb,
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
			Resources: resources,
			Podname:   c.String("pod"),
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
