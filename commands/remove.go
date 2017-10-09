package commands

import (
	"io"

	log "github.com/Sirupsen/logrus"
	pb "github.com/projecteru2/core/rpc/gen"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	cli "gopkg.in/urfave/cli.v2"
)

//RemoveCommand for remove containers
func RemoveCommand() *cli.Command {
	return &cli.Command{
		Name:   "remove",
		Usage:  "remove containers",
		Action: run,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Usage:   "ignore or not before stop hook if it was setted and force check",
				Aliases: []string{"f"},
				Value:   false,
			},
		},
	}
}

func remove(c *cli.Context, conn *grpc.ClientConn) {
	if c.NArg() == 0 {
		log.Fatal("[Remove] not specify containers")
	}
	client := pb.NewCoreRPCClient(conn)
	opts := &pb.RemoveContainerOptions{Ids: c.Args().Slice(), Force: c.Bool("force")}

	resp, err := client.RemoveContainer(context.Background(), opts)
	if err != nil {
		log.Fatalf("[Remove] send request failed %v", err)
	}
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("[Remove] Message invalid %v", err)
		}

		if msg.Success {
			log.Infof("[Remove] Success %s", msg.Id[:12])
			log.Info(msg.Message)
		} else {
			log.Errorf("[Remove] Failed %s", msg.Message)
		}
	}
}
