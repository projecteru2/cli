package main

import (
	"fmt"
	"os"

	"github.com/projecteru2/cli/commands"
	"github.com/projecteru2/cli/versioninfo"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
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
		commands.PublishCommand(),
	}

	app.Flags = commands.GlobalFlags()
	logrus.SetOutput(os.Stdout)
	_ = app.Run(os.Args)
}
