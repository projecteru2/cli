package core

import (
	"context"
	"errors"
	"fmt"
	"io"

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
	for {
		msg, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		for id, addr := range msg.Addresses {
			fmt.Printf("%v: %v\n", id, addr)
		}
	}
}
