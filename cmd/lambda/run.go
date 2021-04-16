package lambda

import (
	"context"
	"fmt"
	"strings"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/interactive"
	corepb "github.com/projecteru2/core/rpc/gen"
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

	memRequest, err := utils.ParseRAMInHuman(c.String("memory-request"))
	if err != nil {
		return nil, fmt.Errorf("[Lambda] memory wrong %v", err)
	}
	memLimit, err := utils.ParseRAMInHuman(c.String("memory"))
	if err != nil {
		return nil, fmt.Errorf("[Lambda] memory wrong %v", err)
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
			ResourceOpts: &corepb.ResourceOptions{
				CpuQuotaRequest: c.Float64("cpu-request"),
				CpuQuotaLimit:   c.Float64("cpu"),
				MemoryRequest:   memRequest,
				MemoryLimit:     memLimit,
				StorageRequest:  c.Int64("storage-request"),
				StorageLimit:    c.Int64("storage"),
				VolumesRequest:  c.StringSlice("volume-request"),
				VolumesLimit:    c.StringSlice("volume"),
			},
			Podname: c.String("pod"),
			NodeFilter: &corepb.NodeFilter{
				Includes: c.StringSlice("node"),
			},
			Image:          c.String("image"),
			Count:          int32(c.Int("count")),
			Env:            c.StringSlice("env"),
			Networks:       utils.GetNetworks(network),
			OpenStdin:      c.Bool("stdin"),
			DeployStrategy: corepb.DeployOptions_Strategy(corepb.DeployOptions_Strategy_value[strings.ToUpper(c.String("deploy-strategy"))]),
			Data:           utils.ReadAllFiles(c.StringSlice("file")),
			User:           c.String("user"),
		},
	}, nil
}
