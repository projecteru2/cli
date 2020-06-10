package commands

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/projecteru2/core/rpc/gen"
	coreutils "github.com/projecteru2/core/utils"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"golang.org/x/net/context"
)

func status(c *cli.Context) error {
	client := setupAndGetGRPCConnection().GetRPCClient()
	name := c.Args().First()
	entry := c.String("entry")
	node := c.String("node")
	labels := makeLabels(c.StringSlice("label"))
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	resp, err := client.ContainerStatusStream(
		ctx,
		&pb.ContainerStatusStreamOptions{
			Appname:    name,
			Entrypoint: entry,
			Nodename:   node,
			Labels:     labels,
		})
	if err != nil || resp == nil {
		return cli.Exit("", -1)
	}

	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil || msg == nil {
			return cli.Exit("", -1)
		}

		if msg.Error != "" {
			if msg.Delete {
				log.Warnf("%s deleted", coreutils.ShortID(msg.Id))
			} else {
				log.Errorf("[%s] status changed with error %v", coreutils.ShortID(msg.Id), msg.Error)
			}
			continue
		}

		if msg.Delete {
			log.Warnf("[%s] %s status expired", coreutils.ShortID(msg.Id), msg.Container.Name)
		}

		if !msg.Status.Running {
			log.Warnf("[%s] %s on %s is stopped", coreutils.ShortID(msg.Id), msg.Container.Name, msg.Container.Nodename)
		} else if !msg.Status.Healthy {
			log.Warnf("[%s] %s on %s is unhealthy", coreutils.ShortID(msg.Id), msg.Container.Name, msg.Container.Nodename)
		} else if msg.Status.Running && msg.Status.Healthy {
			log.Infof("[%s] %s back to life", coreutils.ShortID(msg.Container.Id), msg.Container.Name)
			for networkName, addrs := range msg.Container.Publish {
				log.Infof("[%s] published at %s bind %v", coreutils.ShortID(msg.Id), networkName, addrs)
			}
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
