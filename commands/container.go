package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/projecteru2/core/cluster"
	pb "github.com/projecteru2/core/rpc/gen"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	cli "gopkg.in/urfave/cli.v2"
)

//ContainerCommand for control containers
func ContainerCommand() *cli.Command {
	return &cli.Command{
		Name:  "container",
		Usage: "container commands",
		Subcommands: []*cli.Command{
			&cli.Command{
				Name:      "get",
				Usage:     "get container(s)",
				ArgsUsage: containerArgsUsage,
				Action:    getContainers,
			},
			&cli.Command{
				Name:      "list",
				Usage:     "list container(s) by appname",
				ArgsUsage: "[appname]",
				Action:    listContainers,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "entry",
						Usage: "filter by entry",
					},
					&cli.StringFlag{
						Name:  "nodename",
						Usage: "filter by nodename",
					},
				},
			},
			&cli.Command{
				Name:      "remove",
				Usage:     "remove container(s)",
				ArgsUsage: containerArgsUsage,
				Action:    removeContainers,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Usage:   "ignore or not before stop hook if it was setted and force check",
						Aliases: []string{"f"},
						Value:   false,
					},
				},
			},
			&cli.Command{
				Name:      "copy",
				Usage:     "copy container(s)",
				ArgsUsage: copyArgsUsage,
				Action:    copyContainers,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "dir",
						Usage:   "where to store",
						Aliases: []string{"d"},
						Value:   "/tmp",
					},
				},
			},
			&cli.Command{
				Name:      "realloc",
				Usage:     "realloc containers resource",
				ArgsUsage: containerArgsUsage,
				Action:    reallocContainers,
				Flags: []cli.Flag{
					&cli.Float64Flag{
						Name:    "cpu",
						Usage:   "cpu increment/decrement",
						Aliases: []string{"c"},
						Value:   1.0,
					},
					&cli.Int64Flag{
						Name:    "mem",
						Usage:   "memory increment/decrement",
						Aliases: []string{"m"},
						Value:   134217728,
					},
				},
			},
			&cli.Command{
				Name:      "deploy",
				Usage:     "deploy containers by a image",
				ArgsUsage: specFileURI,
				Action:    deployContainers,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "pod",
						Usage: "where to run",
					},
					&cli.StringFlag{
						Name:  "entry",
						Usage: "which entry",
					},
					&cli.StringFlag{
						Name:  "image",
						Usage: "which to run",
					},
					&cli.StringFlag{
						Name:  "node",
						Usage: "which node to run",
						Value: "",
					},
					&cli.IntFlag{
						Name:  "count",
						Usage: "how many",
						Value: 1,
					},
					&cli.StringFlag{
						Name:  "network",
						Usage: "SDN name or host mode",
						Value: "host",
					},
					&cli.Float64Flag{
						Name:  "cpu",
						Usage: "how many cpu",
						Value: 1.0,
					},
					&cli.Int64Flag{
						Name:  "mem",
						Usage: "how many memory in bytes",
						Value: 536870912.0,
					},
					&cli.StringSliceFlag{
						Name:  "env",
						Usage: "set env can use multiple times",
					},
					&cli.StringSliceFlag{
						Name:  "nodelabel",
						Usage: "set node label can use multiple times",
					},
					&cli.BoolFlag{
						Name:  "with-resource",
						Usage: "resource out control",
						Value: false,
					},
					&cli.StringFlag{
						Name:  "deploy-method",
						Usage: "deploy method auto/fill/each",
						Value: cluster.DeployAuto,
					},
				},
			},
		},
	}
}

func removeContainers(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	opts := &pb.RemoveContainerOptions{Ids: c.Args().Slice(), Force: c.Bool("force")}

	resp, err := client.RemoveContainer(context.Background(), opts)
	if err != nil {
		return cli.Exit(err, -1)
	}
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return cli.Exit(err, -1)
		}

		if msg.Success {
			log.Infof("[RemoveContainer] Success %s", msg.Id[:12])
			if msg.Message != "" {
				log.Info(msg.Message)
			}
		} else {
			log.Errorf("[RemoveContainer] Failed %s", msg.Message)
		}
	}
	return nil
}

func getContainers(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	resp, err := client.GetContainers(context.Background(), &pb.ContainerIDs{Ids: c.Args().Slice()})
	if err != nil {
		return cli.Exit(err, -1)
	}

	for _, container := range resp.GetContainers() {
		log.Infof("ID: %s, Name: %s, Pod: %s, Node: %s", container.GetId(), container.GetName(), container.GetPodname(), container.GetNodename())
	}
	return nil
}

func listContainers(c *cli.Context) error {
	conn := setupAndGetGRPCConnection()
	client := pb.NewCoreRPCClient(conn)

	opts := &pb.DeployStatusOptions{
		Appname:    c.Args().First(),
		Entrypoint: c.String("entry"),
		Nodename:   c.String("nodename"),
	}

	resp, err := client.ListContainers(context.Background(), opts)
	if err != nil {
		return cli.Exit(err, -1)
	}
	for _, container := range resp.Containers {
		log.Infof("%s: %s", container.Name, container.Id)
	}
	return nil
}

func reallocContainers(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	opts := &pb.ReallocOptions{Ids: c.Args().Slice(), Cpu: c.Float64("cpu"), Mem: c.Int64("mem")}

	resp, err := client.ReallocResource(context.Background(), opts)
	if err != nil {
		return cli.Exit(err, -1)
	}
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return cli.Exit(err, -1)
		}

		if msg.Success {
			log.Infof("[Realloc] Success %s", msg.Id[:12])
		} else {
			log.Errorf("[Realloc] Failed %s", msg.Id[:12])
		}
	}
	return nil
}

func copyContainers(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}

	targets := map[string]*pb.CopyPaths{}
	for _, idSrc := range c.Args().Slice() {
		parts := strings.Split(idSrc, ":")
		paths := strings.Split(parts[1], ",")
		targets[parts[0]] = &pb.CopyPaths{Paths: paths}
	}

	opts := &pb.CopyOptions{Targets: targets}
	resp, err := client.Copy(context.Background(), opts)
	if err != nil {
		return cli.Exit(err, -1)
	}

	now := time.Now().Format("2006.01.02.15.04.05")
	baseDir := filepath.Join(c.String("dir"))
	err = os.MkdirAll(baseDir, os.FileMode(0700)) // drwx------
	if err != nil {
		return cli.Exit(err, -1)
	}

	files := map[string]*os.File{}
	defer func() {
		//Close files
		for _, f := range files {
			f.Close()
		}
	}()
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cli.Exit(err, -1)
		}

		if msg.Error != "" {
			log.Errorf("[Copy] Failed %s %s", msg.Id[:12], msg.Error)
			continue
		}

		filename := fmt.Sprintf("%s-%s-%s.tar.gz", msg.Id[:12], msg.Name, now)
		storePath := filepath.Join(baseDir, filename)
		if _, err := os.Stat(storePath); err != nil {
			file, err := os.Create(storePath)
			if err != nil {
				log.Errorf("[Copy] Error during create backup file %s: %v", storePath, err)
				continue
			}
			files[storePath] = file
		}

		_, err = files[storePath].Write(msg.Data)
		if err != nil {
			log.Errorf("[Copy] Write file error %v", err)
		}
	}
	return nil
}
