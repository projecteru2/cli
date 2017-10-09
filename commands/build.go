package commands

import (
	"io"
	"io/ioutil"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/projecteru2/cli/utils"
	pb "github.com/projecteru2/core/rpc/gen"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	cli "gopkg.in/urfave/cli.v2"
	yaml "gopkg.in/yaml.v2"
)

//BuildCommand for building image by multiple stages
func BuildCommand() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "build a image",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "name of image",
			},
			&cli.StringFlag{
				Name:  "tag",
				Usage: "tag of image",
				Value: "latest",
			},
			&cli.StringFlag{
				Name:        "user",
				Usage:       "user of image",
				Value:       "",
				DefaultText: "root",
			},
			&cli.IntFlag{
				Name:        "uid",
				Usage:       "uid of image",
				Value:       0,
				DefaultText: "1",
			},
		},
		Action: run,
	}
}

func build(c *cli.Context, conn *grpc.ClientConn) {
	opts := generateBuildOpts(c)
	client := pb.NewCoreRPCClient(conn)
	resp, err := client.BuildImage(context.Background(), opts)
	if err != nil {
		log.Fatalf("[Build] send request failed %v", err)
	}
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("[Build] Message invalid %v", err)
		}

		if msg.Error != "" {
			log.Errorf("[Build] Error %d %s", msg.ErrorDetail.Code, msg.ErrorDetail.Message)
		} else {
			log.Infof("[Build] %s %s %s", msg.Status, msg.Stream, msg.Progress)

		}
	}
}

func generateBuildOpts(c *cli.Context) *pb.BuildImageOptions {
	if c.NArg() != 1 {
		log.Fatal("[Build] no spec")
	}
	specURI := c.Args().First()
	log.Debugf("[Build] Deploy %s", specURI)
	var data []byte
	var err error
	if strings.HasPrefix(specURI, "http") {
		data, err = utils.GetSpecFromRemote(specURI)
	} else {
		data, err = ioutil.ReadFile(specURI)
	}
	if err != nil {
		log.Fatalf("[Build] read spec failed %v", err)
	}
	specs := &pb.Builds{}
	if err = yaml.Unmarshal(data, specs); err != nil {
		log.Fatalf("[Build] unmarshal specs failed %v", err)
	}

	name := c.String("name")
	user := c.String("user")
	uid := int32(c.Int("uid"))
	tag := c.String("tag")

	opts := &pb.BuildImageOptions{
		Name:   name,
		User:   user,
		Uid:    uid,
		Tag:    tag,
		Builds: specs,
	}
	return opts
}
