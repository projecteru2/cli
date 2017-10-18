package main

import (
	"fmt"
	"os"

	"github.com/projecteru2/cli/commands"
	"github.com/projecteru2/cli/versioninfo"
	"gopkg.in/urfave/cli.v2"
)

func init() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Print(versioninfo.VersionString())
	}
}

func main() {
	app := cli.App{}
	app.Name = versioninfo.NAME
	app.Usage = "control eru in shell"
	app.Version = versioninfo.VERSION

	app.Commands = []*cli.Command{
		commands.ContainerCommand(),
		commands.PodCommand(),
		commands.NodeCommand(),
		commands.ImageCommand(),
		commands.LambdaCommand(),
		commands.StatusCommand(),
	}

	app.Flags = commands.GlobalFlags()
	app.Run(os.Args)
}
