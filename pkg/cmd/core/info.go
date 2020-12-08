package core

import (
	"context"
	"fmt"

	"github.com/projecteru2/cli/pkg/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
)

type coreInfoOptions struct {
	client corepb.CoreRPCClient
}

func (o *coreInfoOptions) run(ctx context.Context) error {
	info, err := o.client.Info(ctx, &corepb.Empty{})
	if err != nil {
		return err
	}

	fmt.Printf("Version:        %s\n", info.Version)
	fmt.Printf("Git hash:       %s\n", info.Revison)
	fmt.Printf("Built:          %s\n", info.BuildAt)
	fmt.Printf("Golang version: %s\n", info.GolangVersion)
	fmt.Printf("OS/Arch:        %s\n", info.OsArch)
	return nil
}

func cmdCoreInfo(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	o := &coreInfoOptions{
		client: client,
	}
	return o.run(c.Context)
}
