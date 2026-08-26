package node

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
)

type addNodeOptions struct {
	client corepb.CoreRPCClient
	opts   *corepb.AddNodeOptions
}

func (o *addNodeOptions) run(ctx context.Context) error {
	node, err := o.client.AddNode(ctx, o.opts)
	if err != nil {
		return err
	}

	describe.Nodes(describe.ToChan(node), false, false)
	return nil
}

func cmdNodeAdd(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	opts, err := generateAddNodeOptions(cmd)
	if err != nil {
		return err
	}

	o := &addNodeOptions{
		client: client,
		opts:   opts,
	}
	return o.run(ctx)
}

func generateAddNodeOptions(cmd *cli.Command) (*corepb.AddNodeOptions, error) {
	podname := cmd.Args().First()
	if podname == "" {
		return nil, errors.New("podname must not be empty")
	}

	nodename := cmd.String("nodename")

	endpoint := cmd.String("endpoint")
	if !strings.Contains(endpoint, "://") {
		return nil, fmt.Errorf("--endpoint must be scheme://address, got %q", endpoint)
	}

	cpumem, storage := collectResourceParams(cmd)
	if cmd.IsSet("cpu") {
		cpumem["cpu"] = cmd.Int("cpu")
	}
	if cmd.IsSet("share") {
		cpumem["share"] = strconv.Itoa(cmd.Int("share"))
	}

	resources, err := utils.EncodeResources(cmd, resourcetypes.Resources{
		utils.ResourceCPUMem:  cpumem,
		utils.ResourceStorage: storage,
	})
	if err != nil {
		return nil, err
	}

	labels := utils.SplitEquality(cmd.StringSlice(flagLabel))
	return &corepb.AddNodeOptions{
		Nodename:  nodename,
		Endpoint:  endpoint,
		Podname:   podname,
		Labels:    labels,
		Resources: resources,
		Test:      cmd.Bool("test"),
	}, nil
}
