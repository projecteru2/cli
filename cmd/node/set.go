package node

import (
	"context"
	"strings"

	"github.com/juju/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
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

	var (
		ca, cert, key string
		err           error
	)
	ca, cert, key, err = readTLSConfigs(c)
	if err != nil {
		return nil, err
	}

	resourceOpts := map[string]*corepb.RawParam{}
	if c.IsSet("cpu") {
		resourceOpts["cpu"] = utils.ToPBRawParamsString(c.String("cpu"))
	}
	if c.IsSet("memory") {
		resourceOpts["memory"] = utils.ToPBRawParamsString(c.String("memory"))
	}
	if c.IsSet("storage") {
		resourceOpts["storage"] = utils.ToPBRawParamsString(c.String("storage"))
	}
	if c.IsSet("volume") {
		resourceOpts["volumes"] = utils.ToPBRawParamsStringSlice(strings.Split(c.String("volume"), ","))
	}
	if c.IsSet("numa") {
		resourceOpts["numa-cpu"] = utils.ToPBRawParamsStringSlice(c.StringSlice("numa"))
	}
	if c.IsSet("numa-memory") {
		resourceOpts["numa-memory"] = utils.ToPBRawParamsStringSlice(c.StringSlice("numa-memory"))
	}
	if c.IsSet("disk") {
		resourceOpts["disks"] = utils.ToPBRawParamsStringSlice(c.StringSlice("disk"))
	}
	if c.IsSet("node-storage-usage-threshold") {
		resourceOpts["node-storage-usage-threshold"] = utils.ToPBRawParamsString(c.Float64("node-storage-usage-threshold"))
	}
	if c.IsSet("pod-storage-usage-threshold") {
		resourceOpts["pod-storage-usage-threshold"] = utils.ToPBRawParamsString(c.Float64("pod-storage-usage-threshold"))
	}
	if c.IsSet("rm-disk") {
		resourceOpts["rm-disks"] = utils.ToPBRawParamsString(c.String("rm-disk"))
	}
	if c.IsSet("workload-limit") {
		resourceOpts["workload-limit"] = utils.ToPBRawParamsStringSlice(c.StringSlice("workload-limit"))
	}
	if c.IsSet("pod-workload-limit") {
		resourceOpts["pod-workload-limit"] = utils.ToPBRawParamsStringSlice(c.StringSlice("pod-workload-limit"))
	}

	return &corepb.SetNodeOptions{
		Nodename:      name,
		ResourceOpts:  resourceOpts,
		Labels:        utils.SplitEquality(c.StringSlice("label")),
		WorkloadsDown: c.Bool("mark-workloads-down"),
		Endpoint:      c.String("endpoint"),
		Delta:         c.Bool("delta"),
		Ca:            ca,
		Cert:          cert,
		Key:           key,
	}, nil
}
