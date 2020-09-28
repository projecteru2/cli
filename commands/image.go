package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/projecteru2/cli/utils"
	dockerengine "github.com/projecteru2/core/engine/docker"
	pb "github.com/projecteru2/core/rpc/gen"
	"github.com/sethgrid/curse"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"golang.org/x/net/context"
	yaml "gopkg.in/yaml.v2"
)

// ImageCommand for building image by multiple stages
func ImageCommand() *cli.Command {
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
				Action: buildImage,
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
				Action: cacheImage,
			},
			{
				Name:      "remove",
				Usage:     "remove image",
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
					&cli.BoolFlag{
						Name:  "prune",
						Usage: "prune node",
						Value: false,
					},
				},
				Action: cleanImage,
			},
		},
	}
}

func buildImage(c *cli.Context) error {
	opts := generateBuildOpts(c)
	client := setupAndGetGRPCConnection(c.Context).GetRPCClient()
	resp, err := client.BuildImage(context.Background(), opts)
	if err != nil {
		return cli.Exit(err, -1)
	}

	progess := map[string]int{}
	p := 0
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return cli.Exit(err, -1)
		}

		if msg.Error != "" { // nolint
			return cli.Exit(msg.ErrorDetail.Message, int(msg.ErrorDetail.Code))
		} else if msg.Stream != "" {
			fmt.Print(msg.Stream)
			if msg.Status == "finished" {
				progess = map[string]int{}
				p = 0
			}
		} else if msg.Status != "" {
			if msg.Id == "" {
				fmt.Println(msg.Status)
			} else {
				data := fmt.Sprintf("%s: %s %s", msg.Id, msg.Status, msg.Progress)
				if pos, ok := progess[msg.Id]; !ok {
					progess[msg.Id] = p
					fmt.Println(data)
					p++
				} else {
					cursor, err := curse.New()
					if err != nil {
						//log.Errorf("[Build] get cursor failed %v", err)
						fmt.Print(data)
						continue
					}
					cursor.MoveUp(p - pos).EraseCurrentLine()
					fmt.Print(data)
					cursor.Reset()
				}
			}
		}
	}
	return nil
}

func cacheImage(c *cli.Context) error {
	opts := &pb.CacheImageOptions{
		Images:    c.Args().Slice(),
		Step:      int32(c.Int("concurrent")),
		Podname:   c.String("podname"),
		Nodenames: c.StringSlice("nodename"),
	}

	client := setupAndGetGRPCConnection(c.Context).GetRPCClient()
	resp, err := client.CacheImage(context.Background(), opts)
	if err != nil {
		return cli.Exit(err, -1)
	}

	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return cli.Exit(err, -1)
		}

		if msg.Success {
			log.Infof("[CacheImage] cache image %s on %s success", msg.Image, msg.Nodename)
		} else {
			log.Warnf("[CacheImage] cache image %s on %s failed", msg.Image, msg.Nodename)
		}
	}
	return nil
}

func cleanImage(c *cli.Context) error {
	opts := &pb.RemoveImageOptions{
		Images:    c.Args().Slice(),
		Step:      int32(c.Int("concurrent")),
		Podname:   c.String("podname"),
		Nodenames: c.StringSlice("nodename"),
		Prune:     c.Bool("prune"),
	}
	client := setupAndGetGRPCConnection(c.Context).GetRPCClient()
	resp, err := client.RemoveImage(context.Background(), opts)
	if err != nil {
		return cli.Exit(err, -1)
	}

	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return cli.Exit(err, -1)
		}

		if msg.Success {
			log.Infof("[CleanImage] Success remove %s", msg.Image)
		} else {
			log.Errorf("[Cleanimage] Failed remove %s", msg.Image)
		}
		for _, m := range msg.Messages {
			log.Infof(m)
		}
	}

	return nil
}

func generateBuildOpts(c *cli.Context) *pb.BuildImageOptions {
	if c.NArg() != 1 {
		log.Fatal("[Build] no spec")
	}
	raw := c.Bool("raw")
	exist := c.Bool("exist")
	if raw && exist {
		log.Fatal("[Build] mutually exclusive flag: raw or exist")
	}
	stopSignal := c.String("stop-signal")

	var specs *pb.Builds
	var tar []byte
	var existID string
	var buildMethod pb.BuildImageOptions_BuildMethod
	switch {
	case exist:
		buildMethod = pb.BuildImageOptions_EXIST
		existID = c.Args().First()
	case !raw:
		buildMethod = pb.BuildImageOptions_SCM
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
		data, err = utils.EnvParser(data)
		if err != nil {
			log.Fatalf("[Build] parse env failed %v", err)
		}
		specs = &pb.Builds{}
		if err = yaml.Unmarshal(data, specs); err != nil {
			log.Fatalf("[Build] unmarshal specs failed %v", err)
		}
		for s := range specs.Builds {
			b := specs.Builds[s]
			b.StopSignal = stopSignal
		}
	default:
		buildMethod = pb.BuildImageOptions_RAW
		path := c.Args().First()
		data, err := dockerengine.CreateTarStream(path)
		if err != nil {
			log.Fatal("[Build] no path")
		}
		tar, err = ioutil.ReadAll(data)
		if err != nil {
			log.Fatal("[Build] create tar stream failed")
		}
	}

	name := c.String("name")
	if name == "" {
		log.Fatal("[Build] need name")
	}
	user := c.String("user")
	uid := int32(c.Int("uid"))
	tags := c.StringSlice("tag")
	if len(tags) == 0 {
		tags = append(tags, "latest")
	}

	opts := &pb.BuildImageOptions{
		Name:        name,
		User:        user,
		Uid:         uid,
		Tags:        tags,
		BuildMethod: buildMethod,
		Builds:      specs,
		Tar:         tar,
		ExistId:     existID,
	}
	return opts
}
