package commands

import (
	"encoding/json"
	"io"

	pb "github.com/projecteru2/core/rpc/gen"
	"github.com/projecteru2/core/store"
	coretypes "github.com/projecteru2/core/types"
	coreutils "github.com/projecteru2/core/utils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	cli "github.com/urfave/cli/v2"
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
	if err != nil || resp == nil {
		cli.Exit("", -1)
	}

	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil || msg == nil {
			cli.Exit("", -1)
		}

		if msg.Action == store.DeleteEvent {
			log.Infof("[%s] %s_%s deleted", coreutils.ShortID(msg.Id), msg.Appname, msg.Entrypoint)
			continue
		}

		meta := &coretypes.Meta{}
		if len(msg.Data) > 0 {
			if err := json.Unmarshal(msg.Data, meta); err != nil {
				log.Errorf("[status] parse container data failed %v", err)
				break
			}
		}

		if !coreutils.FilterContainer(meta.Labels, labels) {
			log.Debugf("[status] ignore container %s", msg.Id)
			continue
		}

		if meta.Healthy {
			log.Infof("[%s] %s_%s on %s back to life", coreutils.ShortID(msg.Id), msg.Appname, msg.Entrypoint, msg.Nodename)
			containerMeta := coreutils.DecodeMetaInLabel(meta.Labels)
			publish := coreutils.MakePublishInfo(meta.Networks, containerMeta.Publish)
			for networkName, addrs := range publish {
				log.Infof("[%s] published at %s bind %v", coreutils.ShortID(msg.Id), networkName, addrs)
			}
		} else if !meta.Running {
			log.Warnf("[%s] %s_%s on %s is stopped", coreutils.ShortID(msg.Id), msg.Appname, msg.Entrypoint, msg.Nodename)
		} else if !meta.Healthy {
			log.Warnf("[%s] %s_%s on %s is unhealthy", coreutils.ShortID(msg.Id), msg.Appname, msg.Entrypoint, msg.Nodename)
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
