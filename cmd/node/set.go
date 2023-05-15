package node

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/juju/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/projecteru2/cli/cmd/utils"
	resourcetypes "github.com/projecteru2/core/resource/types"
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

	cpumem := resourcetypes.RawParams{}
	storage := resourcetypes.RawParams{}

	if c.IsSet("cpu") {
		cpumem["cpu"] = c.String("cpu")
	}
	if c.IsSet("share") {
		cpumem["share"] = c.String("share")
	}
	if c.IsSet("memory") {
		cpumem["memory"] = c.String("memory")
	}
	if c.IsSet("numa-cpu") {
		cpumem["numa-cpu"] = c.StringSlice("numa-cpu")
	}
	if c.IsSet("numa-memory") {
		cpumem["numa-memory"] = c.StringSlice("numa-memory")
	}
	if c.IsSet("disk") {
		storage["disks"] = c.StringSlice("disk")
	}
	if c.IsSet("storage") {
		storage["storage"] = c.String("storage")
	}
	if c.IsSet("volume") {
		storage["volumes"] = c.StringSlice("volume")
	}
	if c.IsSet("rm-disk") {
		storage["rm-disks"] = c.String("rm-disk")
	}

	cb, _ := json.Marshal(cpumem)
	sb, _ := json.Marshal(storage)
	resources := map[string][]byte{
		"cpumem":  cb,
		"storage": sb,
	}

	if extraResourcesMap, err := utils.ParseExtraResources(c); err != nil {
		for k, v := range extraResourcesMap {
			if _, ok := resources[k]; ok {
				continue
			}
			eb, _ := json.Marshal(v)
			resources[k] = eb
		}
	} else {
		return nil, fmt.Errorf("[generateSetNodeOptions] get extra resources failed %v", err)
	}

	return &corepb.SetNodeOptions{
		Nodename:      name,
		Resources:     resources,
		Labels:        utils.SplitEquality(c.StringSlice("label")),
		WorkloadsDown: c.Bool("mark-workloads-down"),
		Endpoint:      c.String("endpoint"),
		Delta:         c.Bool("delta"),
		Ca:            ca,
		Cert:          cert,
		Key:           key,
	}, nil
}
