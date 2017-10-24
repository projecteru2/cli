package commands

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/projecteru2/cli/utils"
	pb "github.com/projecteru2/core/rpc/gen"
	"google.golang.org/grpc"
	cli "gopkg.in/urfave/cli.v2"
)

var (
	debug bool
	eru   string
)

const (
	containerArgsUsage = "containerID(s)"
	podArgsUsage       = "podname"
	nodeArgsUsage      = "nodename"
	specFileURI        = "<spec file uri>"
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
	}
}

func setupAndGetGRPCConnection() *grpc.ClientConn {
	setupLog("INFO")
	if debug {
		setupLog("DEBUG")
	}
	return utils.ConnectEru(eru)
}

func checkParamsAndGetClient(c *cli.Context) (pb.CoreRPCClient, error) {
	if c.NArg() == 0 {
		return nil, fmt.Errorf("not specify arguments")
	}
	conn := setupAndGetGRPCConnection()
	client := pb.NewCoreRPCClient(conn)
	return client, nil
}
