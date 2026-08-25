package status

import (
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

// Command returns the status command tree.
func Command() *cli.Command {
	return &cli.Command{
		Name:      "status",
		Usage:     "get deploy status from core",
		ArgsUsage: "name can be none",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "entry",
				Usage: "entry filter or not",
			},
			&cli.StringFlag{
				Name:  "node",
				Usage: "node filter or not",
			},
			&cli.StringSliceFlag{
				Name:  "label",
				Usage: "label filter can set multiple times",
			},
		},
		Action: utils.ExitCoder(cmdStatus),
	}
}
