package commands

import (
	"fmt"

	pb "github.com/projecteru2/core/rpc/gen"
	cli "github.com/urfave/cli/v2"
)

// CoreCommand for core
func CoreCommand() *cli.Command {
	return &cli.Command{
		Name:  "core",
		Usage: "core commands",
		Subcommands: []*cli.Command{
			{
				Name:   "info",
				Usage:  "core info",
				Action: info,
			},
		},
	}
}

func info(c *cli.Context) error {
	client := setupAndGetGRPCConnection(c.Context).GetRPCClient()
	opts := &pb.Empty{}
	coreInfo, err := client.Info(c.Context, opts)
	if err != nil {
		return cli.Exit(err, -1)
	}
	fmt.Printf("Version:        %s\n", coreInfo.Version)
	fmt.Printf("Git hash:       %s\n", coreInfo.Revison)
	fmt.Printf("Built:          %s\n", coreInfo.BuildAt)
	fmt.Printf("Golang version: %s\n", coreInfo.GolangVersion)
	fmt.Printf("OS/Arch:        %s\n", coreInfo.OsArch)
	return nil
}
