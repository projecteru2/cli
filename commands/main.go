package commands

import (
	"context"
	"fmt"

	"github.com/projecteru2/cli/utils"
	pb "github.com/projecteru2/core/rpc/gen"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	cli "gopkg.in/urfave/cli.v2"
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
	}
}

func setupAndGetGRPCConnection() *grpc.ClientConn {
	setupLog("INFO")
	if debug {
		setupLog("DEBUG")
	}
	opts := []grpc.DialOption{grpc.WithInsecure()}
	if username != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(new(basicCredential)))
	}
	return utils.ConnectEru(eru, opts)
}

func checkParamsAndGetClient(c *cli.Context) (pb.CoreRPCClient, error) {
	if c.NArg() == 0 {
		return nil, fmt.Errorf("not specify arguments")
	}
	conn := setupAndGetGRPCConnection()
	client := pb.NewCoreRPCClient(conn)
	return client, nil
}

// customCredential 自定义认证
type basicCredential struct{}

func (c basicCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		username: password,
	}, nil
}

func (c basicCredential) RequireTransportSecurity() bool {
	return false
}
