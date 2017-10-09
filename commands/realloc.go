package commands

import (
	"io"

	log "github.com/Sirupsen/logrus"
	pb "github.com/projecteru2/core/rpc/gen"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	cli "gopkg.in/urfave/cli.v2"
)

//ReallocCommand for realloc containers resource
func ReallocCommand() *cli.Command {
	return &cli.Command{
		Name:   "realloc",
		Usage:  "realloc containers resource",
		Action: run,
		Flags: []cli.Flag{
			&cli.Float64Flag{
				Name:    "cpu",
				Usage:   "cpu increment/decrement",
				Aliases: []string{"c"},
				Value:   1.0,
			},
			&cli.Int64Flag{
				Name:    "mem",
				Usage:   "memory increment/decrement",
				Aliases: []string{"m"},
				Value:   134217728,
			},
		},
	}
}

func realloc(c *cli.Context, conn *grpc.ClientConn) {
	if c.NArg() == 0 {
		log.Fatal("[Realloc] not specify containers")
	}
	client := pb.NewCoreRPCClient(conn)
	opts := &pb.ReallocOptions{Ids: c.Args().Slice(), Cpu: c.Float64("cpu"), Mem: c.Int64("mem")}

	resp, err := client.ReallocResource(context.Background(), opts)
	if err != nil {
		log.Fatalf("[Realloc] send request failed %v", err)
	}
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("[Realloc] Message invalid %v", err)
		}

		if msg.Success {
			log.Infof("[Realloc] Success %s", msg.Id[:12])
		} else {
			log.Errorf("[Realloc] Failed %s", msg.Id[:12])
		}
	}
}
