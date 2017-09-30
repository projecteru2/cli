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

	deployCommand := commands.DeployCommand()
	removeCommand := commands.RemoveCommand()

	app.Commands = []*cli.Command{
		deployCommand,
		removeCommand,
	}

	app.Flags = commands.GlobalFlags()
	app.Run(os.Args)
}
