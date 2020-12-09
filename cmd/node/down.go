package node

import (
	"context"
	"time"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type setNodeDownOptions struct {
	client       corepb.CoreRPCClient
	name         string
	check        bool
	checkTimeout int
}

func (o *setNodeDownOptions) run(ctx context.Context) error {
	do := true
	if o.check {
		timeout, cancel := context.WithTimeout(ctx, time.Duration(o.checkTimeout)*time.Second)
		defer cancel()
		if _, err := o.client.GetNodeResource(timeout, &corepb.GetNodeResourceOptions{
			Opts: &corepb.GetNodeOptions{
				Nodename: o.name,
			},
		}); err == nil {
			logrus.Warn("[SetNode] node is not down")
			do = false
		}
	}

	if do {
		_, err := o.client.SetNode(ctx, &corepb.SetNodeOptions{
			Nodename:  o.name,
			StatusOpt: corepb.TriOpt_FALSE,
		})
		if err != nil {
			return err
		}
		logrus.Infof("[SetNode] node %s down", o.name)
	}
	return nil
}

func cmdNodeSetDown(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Node name must be given")
	}

	o := &setNodeDownOptions{
		client:       client,
		name:         name,
		check:        c.Bool("check"),
		checkTimeout: c.Int("check-timeout"),
	}
	return o.run(c.Context)
}
