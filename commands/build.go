package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/projecteru2/cli/utils"
	pb "github.com/projecteru2/core/rpc/gen"
	"github.com/sethgrid/curse"
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

type Pos struct {
	Col  int
	Line int
}

func build(c *cli.Context, conn *grpc.ClientConn) {
	opts := generateBuildOpts(c)
	client := pb.NewCoreRPCClient(conn)
	resp, err := client.BuildImage(context.Background(), opts)
	if err != nil {
		log.Fatalf("[Build] send request failed %v", err)
	}

	cursor, err := curse.New()
	if err != nil {
		log.Fatalf("[Build] get cursor failed %v", err)
	}
	newLine := cursor.StartingPosition.Y
	progess := map[string]Pos{}
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("[Build] Message invalid %v", err)
		}

		cursor.Move(0, newLine).EraseCurrentLine()
		if msg.Error != "" {
			fmt.Print(msg.ErrorDetail.Message)
			newLine++
		} else if msg.Stream != "" && msg.Progress == "" {
			fmt.Print(msg.Stream)
			newLine++
		} else if msg.Status != "" {
			if msg.Id == "" {
				fmt.Printf("%s", msg.Status)
				newLine++
				continue
			}
			col, line, err := curse.GetCursorPosition()
			if err != nil {
				log.Fatalf("[Build] get cursor pos failed %v", err)
			}
			if pos, ok := progess[msg.Id]; !ok {
				progess[msg.Id] = Pos{col, line}
				newLine++
			} else {
				cursor.Move(pos.Col, pos.Line).EraseCurrentLine()
			}
			fmt.Printf("%s: %s %s", msg.Id, msg.Status, msg.Progress)
		}
		cursor.Move(0, newLine)
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
