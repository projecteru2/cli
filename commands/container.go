package commands

import (
	log "github.com/Sirupsen/logrus"
	pb "github.com/projecteru2/core/rpc/gen"
	"golang.org/x/net/context"
	cli "gopkg.in/urfave/cli.v2"
)

// Container command
func ContainerCommand() *cli.Command {
	return &cli.Command{
		Name:  "container",
		Usage: "container commands",
		SubCommands: []*cli.Command{
			&cli.Command{
				Name:   "get",
				Usage:  "get container(s)",
				Action: getContainers,
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:  "ids",
						Usage: "id(s) of container",
						Value: &cli.StringSlice{},
					},
				},
			},
		},
	}
}

func getContainers(c *cli.Context) error {
	ids := c.StringSlice("ids")
	conn := setupAndGetGRPCConnection()
	client := pb.NewCoreRPCClient(conn)
	resp, err := client.GetContainers(context.Background(), &pb.ContainerIDs{
		Ids: ids,
	})
	if err != nil {
		log.Fatalf("[GetContainers] send request failed %v", err)
	}

	for _, container := range resp.GetContainers() {
		log.Infof("ID: %s, Name: %s, Pod: %s, Node: %s", container.GetId(), container.GetName(), container.GetPodname(), container.GetNodename())
	}
	return nil
}
