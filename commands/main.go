package commands

import (
	"fmt"

	"github.com/projecteru2/core/client"
	pb "github.com/projecteru2/core/rpc/gen"
	"github.com/projecteru2/core/types"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

var (
	debug    bool
	eru      string
	username string
	password string
)

const (
	containerArgsUsage = "containerID(s)"
	podArgsUsage       = "podname"
	nodeArgsUsage      = "nodename"
	specFileURI        = "<spec file uri>"
	copyArgsUsage      = "containerID:path1,path2,...,pathn"
	sendArgsUsage      = "path1,path2,...pathn"
)

func setupLog(l string) error {
	level, err := log.ParseLevel(l)
	if err != nil {
		return err
	}
	log.SetLevel(level)

	formatter := &log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}
	log.SetFormatter(formatter)
	return nil
}

//GlobalFlags for global control
func GlobalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "debug",
			Usage:       "enable debug",
			Aliases:     []string{"d"},
			Value:       false,
			Destination: &debug,
		},
		&cli.StringFlag{
			Name:        "eru",
			Usage:       "eru core address",
			Aliases:     []string{"e"},
			Value:       "127.0.0.1:5001",
			EnvVars:     []string{"ERU"},
			Destination: &eru,
		},
		&cli.StringFlag{
			Name:        "username",
			Usage:       "eru core username",
			Aliases:     []string{"u"},
			Value:       "",
			EnvVars:     []string{"ERU_USERNAME"},
			Destination: &username,
		},
		&cli.StringFlag{
			Name:        "password",
			Usage:       "eru core password",
			Aliases:     []string{"p"},
			Value:       "",
			EnvVars:     []string{"ERU_PASSWORD"},
			Destination: &password,
		},
		&cli.BoolFlag{
			Name:    "pretty",
			Usage:   "use table to output",
			Value:   false,
			EnvVars: []string{"ERU_PRETTY_PRINT"},
		},
	}
}

func setupAndGetGRPCConnection() *client.Client {
	_ = setupLog("INFO")
	if debug {
		_ = setupLog("DEBUG")
	}

	return client.NewClient(eru, types.AuthConfig{Username: username, Password: password})
}

func checkParamsAndGetClient(c *cli.Context) (pb.CoreRPCClient, error) {
	if c.NArg() == 0 {
		return nil, fmt.Errorf("not specify arguments")
	}
	return setupAndGetGRPCConnection().GetRPCClient(), nil
}
