package image

import (
	"github.com/projecteru2/cli/cmd/utils"
	"github.com/urfave/cli/v2"
)

const (
	specFileURI = "<spec file uri>"
)

// Command exports image subcommands
func Command() *cli.Command {
	return &cli.Command{
		Name:  "image",
		Usage: "image commands",
		Subcommands: []*cli.Command{
			{
				Name:      "build",
				Usage:     "build image",
				ArgsUsage: specFileURI,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "name",
						Usage: "name of image",
					},
					&cli.StringSliceFlag{
						Name:  "tag",
						Usage: "tag of image",
					},
					&cli.BoolFlag{
						Name:  "raw",
						Usage: "build image from dir",
					},
					&cli.BoolFlag{
						Name:  "exist",
						Usage: "build image from exist",
					},
					&cli.StringFlag{
						Name:        "user",
						Usage:       "user of image",
						Value:       "",
						DefaultText: "root",
					},
					&cli.StringFlag{
						Name:  "stop-signal",
						Usage: "customize stop signal",
					},
					&cli.IntFlag{
						Name:        "uid",
						Usage:       "uid of image",
						Value:       0,
						DefaultText: "1",
					},
				},
				Action: utils.ExitCoder(cmdImageBuild),
			},
			{
				Name:      "cache",
				Usage:     "cache image",
				ArgsUsage: "name of images",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:  "nodename",
						Usage: "nodename if you just want to cache on one node",
					},
					&cli.StringFlag{
						Name:  "podname",
						Usage: "name of pod, if you want to cache on all nodes in one pod",
					},
					&cli.IntFlag{
						Name:  "concurrent",
						Usage: "how many workers to pull images",
						Value: 10,
					},
				},
				Action: utils.ExitCoder(cmdImageCache),
			},
			{
				Name:      "remove",
				Usage:     "remove image",
				ArgsUsage: "name of images",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:  "nodename",
						Usage: "nodename if you just want to clean on one node",
					},
					&cli.StringFlag{
						Name:  "podname",
						Usage: "name of pod, if you want to clean on all nodes in one pod",
					},
					&cli.IntFlag{
						Name:  "concurrent",
						Usage: "how many workers to pull images",
						Value: 10,
					},
					&cli.BoolFlag{
						Name:  "prune",
						Usage: "prune node",
						Value: false,
					},
				},
				Action: utils.ExitCoder(cmdImageClean),
			},
		},
	}
}
