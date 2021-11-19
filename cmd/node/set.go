package node

import (
	"context"

	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/juju/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
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
	logrus.Infof("[SetNode] set node %s success", o.opts.Nodename)
	return nil
}

func cmdNodeSet(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	opts, err := generateSetNodeOptions(c, client)
	if err != nil {
		return err
	}

	o := &setNodeOptions{
		client: client,
		opts:   opts,
	}
	return o.run(c.Context)
}

func generateSetNodeOptions(c *cli.Context, _ corepb.CoreRPCClient) (*corepb.SetNodeOptions, error) {
	name := c.Args().First()
	if name == "" {
		return nil, errors.New("Node name must be given")
	}

	stringFlags := []string{"cpu", "memory", "storage", "volume"}
	stringSliceFlags := []string{"numa-cpu", "numa-memory"}
	resourceOpts := utils.GetResourceOpts(c, stringFlags, stringSliceFlags, nil, nil)

	return &corepb.SetNodeOptions{
		Nodename:      name,
		StatusOpt:     corepb.TriOpt_KEEP,
		ResourceOpts:  resourceOpts,
		Labels:        utils.SplitEquality(c.StringSlice("label")),
		WorkloadsDown: c.Bool("mark-workloads-down"),
		Delta:         c.Bool("delta"),
	}, nil
}
