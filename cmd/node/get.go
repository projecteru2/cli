package node

import (
	"context"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type getNodeOptions struct {
	client corepb.CoreRPCClient
	name   string
}

func (o *getNodeOptions) run(ctx context.Context) error {
	node, err := o.client.GetNode(ctx, &corepb.GetNodeOptions{
		Nodename: o.name,
	})
	if err != nil {
		return err
	}

	logrus.Infof("Name: %s, Endpoint: %s", node.GetName(), node.GetEndpoint())
	for k, v := range node.GetLabels() {
		logrus.Infof("%s: %s", k, v)
	}
	logrus.Infof("CPU Used: %.2f", node.GetCpuUsed())
	logrus.Infof("Memory Used: %d/%d bytes", node.GetMemoryUsed(), node.GetInitMemory())
	for nodeID, memory := range node.GetNumaMemory() {
		logrus.Infof("Memory Node: %s Capacity %d bytes", nodeID, memory)
	}

	initVolume := node.GetInitVolume()
	totalCap := int64(0)
	for volume, freeSpace := range node.GetVolume() {
		capacity := initVolume[volume]
		totalCap += capacity
		logrus.Infof("  Volume %s: Used %d/%d bytes", volume, capacity-freeSpace, capacity)
	}
	logrus.Infof("Volume Used: %d/%d bytes", node.GetVolumeUsed(), totalCap)

	if node.GetInitStorage() > 0 {
		logrus.Infof("Storage Used: %d bytes", node.GetStorageUsed())
	} else {
		logrus.Infof("Storage Used: %d bytes (%s)", node.GetStorageUsed(), "UNLIMITED")
	}
	return nil
}

func cmdNodeGet(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Node name must be given")
	}

	o := &getNodeOptions{
		client: client,
		name:   name,
	}
	return o.run(c.Context)
}
