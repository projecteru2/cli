package commands

import (
	log "github.com/Sirupsen/logrus"
	"github.com/projecteru2/cli/utils"
	"google.golang.org/grpc"
	cli "gopkg.in/urfave/cli.v2"
)

var (
	debug   bool
	eru     string
	timeout int
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
		&cli.IntFlag{
			Name:        "timeout",
			Usage:       "timeout for conn",
			Aliases:     []string{"t"},
			Value:       2,
			Destination: &timeout,
		},
	}
}

func setupAndGetGRPCConnection() *grpc.ClientConn {
	setupLog("INFO")
	if debug {
		setupLog("DEBUG")
	}
	return utils.ConnectEru(eru, timeout)
}

func run(c *cli.Context) error {
	conn := setupAndGetGRPCConnection()
	if c.Command.Name == "deploy" {
		deploy(c, conn)
	} else if c.Command.Name == "remove" {
		remove(c, conn)
	} else if c.Command.Name == "realloc" {
		realloc(c, conn)
	} else if c.Command.Name == "build" {
		errDetail := build(c, conn)
		if errDetail != nil {
			return cli.Exit(errDetail.Message, int(errDetail.Code))
		}
	} else if c.Command.Name == "lambda" {
		code := lambda(c, conn)
		if code != 0 {
			return cli.Exit("", code)
		}
	} else {
		log.Fatal("Not support yet")
	}
	return nil
}
