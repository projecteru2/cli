package core

import (
	"context"
	"fmt"

	"github.com/projecteru2/core/log"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

func cmdWatchServiceStatus(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	resp, err := client.WatchServiceStatus(ctx, &corepb.Empty{})
	if err != nil {
		return err
	}
	log.WithFunc("core.cmdWatchServiceStatus").Info(ctx, "watch start")
	return utils.EachMessage(resp.Recv, func(msg *corepb.ServiceStatus) error {
		for id, addr := range msg.Addresses {
			fmt.Printf("%v: %v\n", id, addr)
		}
		return nil
	})
}
