package commands

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/projecteru2/cli/types"
	pb "github.com/projecteru2/core/rpc/gen"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	cli "gopkg.in/urfave/cli.v2"
)

func status(c *cli.Context) error {
	client := setupAndGetGRPCConnection().GetRPCClient()
	name := c.String("name")
	entry := c.String("entry")
	node := c.String("node")
	version := c.String("version")
	extend := c.StringSlice("extend")

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

		if !filterContainer(container, version, extend) {
			log.Debugf("[status] ignore container %v", container)
		}

		if msg.Action == "delete" {
			log.Infof("[%s] %s_%s deleted", container.ID[:12], container.Name, container.EntryPoint)
		} else if msg.Action == "set" || msg.Action == "update" {
			if container.Healthy {
				pub := []string{}
				if container.Healthy {
					for _, addr := range container.Publish {
						pub = append(pub, addr)
					}
				}
				log.Infof("[%s] %s_%s on %s published at %s", container.ID[:12], container.Name, container.EntryPoint, container.Nodename, strings.Join(pub, ","))
			} else {
				log.Warnf("[%s] %s_%s on %s is unhealthy", container.ID[:12], container.Name, container.EntryPoint, container.Nodename)
			}
		}
	}
	return nil
}

func filterContainer(container *types.Container, version string, extend []string) bool {
	if version != "" && container.Version != version {
		return false
	}
	ext := map[string]string{}
	for _, d := range extend {
		p := strings.Split(d, "=")
		ext[p[0]] = ext[p[1]]
	}
	for k, v := range ext {
		if n, ok := container.Extend[k]; !ok || n != v {
			return false
		}
	}
	return true
}

// StatusCommand show status
func StatusCommand() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "get deploy status from core",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "name filter or not",
			},
			&cli.StringFlag{
				Name:  "entry",
				Usage: "entry filter or not",
			},
			&cli.StringFlag{
				Name:  "node",
				Usage: "node filter or not",
			},
			&cli.StringFlag{
				Name:  "version",
				Usage: "version filter or not",
			},
			&cli.StringSliceFlag{
				Name:  "extend",
				Usage: "extend filter can set multiple times",
			},
		},
		Action: status,
	}
}
