package core

import (
	"fmt"
	"io"

	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
)

func cmdWatchServiceStatus(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	resp, err := client.WatchServiceStatus(c.Context, &corepb.Empty{})
	if err != nil {
		return err
	}
	fmt.Println("watch start")
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}
		for id, addr := range msg.Addresses {
			fmt.Printf("%v: %v\n", id, addr)
		}
	}
	return nil
}
