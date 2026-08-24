package core

import (
	"github.com/projecteru2/cli/cmd/utils"

	"github.com/urfave/cli/v3"
)

// Command exports core subommands
func Command() *cli.Command {
	return &cli.Command{
		Name:  "core",
		Usage: "core commands",
		Commands: []*cli.Command{
			{
				Name:   "info",
				Usage:  "core info",
				Action: utils.ExitCoder(cmdCoreInfo),
			},
			{
				Name:   "watch",
				Usage:  "",
				Action: utils.ExitCoder(cmdWatchServiceStatus),
			},
		},
	}
}
