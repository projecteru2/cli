package core

import (
	"github.com/urfave/cli/v2"
)

// Command exports core subommands
func Command() *cli.Command {
	return &cli.Command{
		Name:  "core",
		Usage: "core commands",
		Subcommands: []*cli.Command{
			{
				Name:   "info",
				Usage:  "core info",
				Action: cmdCoreInfo,
			},
		},
	}
}
