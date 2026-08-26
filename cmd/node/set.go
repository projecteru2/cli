package node

import (
	"context"
	"errors"

	"github.com/projecteru2/core/log"
	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type setNodeOptions struct {
	client corepb.CoreRPCClient
	opts   *corepb.SetNodeOptions
}

func (o *setNodeOptions) run(ctx context.Context) error {
	_, err := o.client.SetNode(ctx, o.opts)
	if err != nil {
		return err
	}
	log.WithFunc("node.setNodeOptions.run").Infof(ctx, "set node %s success", o.opts.Nodename)
	return nil
}

func cmdNodeSet(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	opts, err := generateSetNodeOptions(cmd)
	if err != nil {
		return err
	}

	o := &setNodeOptions{
		client: client,
		opts:   opts,
	}
	return o.run(ctx)
}

func generateSetNodeOptions(cmd *cli.Command) (*corepb.SetNodeOptions, error) {
	name := cmd.Args().First()
	if name == "" {
		return nil, errors.New("node name must be given")
	}

	cpumem, storage := collectResourceParams(cmd)
	if cmd.IsSet("cpu") {
		cpumem["cpu"] = cmd.String("cpu")
	}
	if cmd.IsSet("rm-disk") {
		storage["rm-disks"] = cmd.String("rm-disk")
	}

	resources, err := utils.EncodeResources(cmd, resourcetypes.Resources{
		utils.ResourceCPUMem:  cpumem,
		utils.ResourceStorage: storage,
	})
	if err != nil {
		return nil, err
	}

	return &corepb.SetNodeOptions{
		Nodename:      name,
		Resources:     resources,
		Labels:        utils.SplitEquality(cmd.StringSlice(flagLabel)),
		WorkloadsDown: cmd.Bool("mark-workloads-down"),
		Endpoint:      cmd.String("endpoint"),
		Delta:         cmd.Bool("delta"),
	}, nil
}
