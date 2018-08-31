package commands

import (
	"encoding/json"
	"io"

	"github.com/projecteru2/cli/types"
	pb "github.com/projecteru2/core/rpc/gen"
	coreutils "github.com/projecteru2/core/utils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	cli "gopkg.in/urfave/cli.v2"
)

func status(c *cli.Context) error {
	client := setupAndGetGRPCConnection().GetRPCClient()
	name := c.Args().First()
	entry := c.String("entry")
	node := c.String("node")
	labels := makeLabels(c.StringSlice("label"))

	resp, err := client.DeployStatus(
		context.Background(),
		&pb.DeployStatusOptions{
			Appname:    name,
			Entrypoint: entry,
			Nodename:   node,
		})
	if err != nil {
		cli.Exit("", -1)
	}

	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			cli.Exit("", -1)
		}

		container := &types.Container{}
		if len(msg.Data) > 0 {
			if err := json.Unmarshal(msg.Data, container); err != nil {
				log.Errorf("[status] parse container data failed %v", err)
				break
			}
		}
		container.ID = msg.Id
		container.Name = msg.Appname
		container.EntryPoint = msg.Entrypoint
		container.Nodename = msg.Nodename

		if !filterContainer(container.Labels, labels) {
			log.Debugf("[status] ignore container %s", container.ID)
			continue
		}

		if msg.Action == "delete" {
			log.Infof("[%s] %s_%s deleted", coreutils.ShortID(container.ID), container.Name, container.EntryPoint)
		} else if msg.Action == "set" || msg.Action == "update" {
			if container.Healthy {
				log.Infof("[%s] %s_%s on %s back to life", coreutils.ShortID(container.ID), container.Name, container.EntryPoint, container.Nodename)
				for networkName, addrs := range container.Publish {
					log.Infof("[%s] published at %s bind %v", coreutils.ShortID(container.ID), networkName, addrs)
				}
				continue
			}
			log.Warnf("[%s] %s_%s on %s is unhealthy", coreutils.ShortID(container.ID), container.Name, container.EntryPoint, container.Nodename)
		}
	}
	return nil
}

// StatusCommand show status
func StatusCommand() *cli.Command {
	return &cli.Command{
		Name:      "status",
		Usage:     "get deploy status from core",
		ArgsUsage: "name can be none",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "entry",
				Usage: "entry filter or not",
			},
			&cli.StringFlag{
				Name:  "node",
				Usage: "node filter or not",
			},
			&cli.StringSliceFlag{
				Name:  "label",
				Usage: "label filter can set multiple times",
			},
		},
		Action: status,
	}
}
