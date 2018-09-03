package commands

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/projecteru2/cli/utils"
	pb "github.com/projecteru2/core/rpc/gen"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	cli "gopkg.in/urfave/cli.v2"
)

func replaceContainers(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	specURI := c.Args().First()
	log.Debugf("[Replace] Replace container by %s", specURI)

	pod, node, entry, image, network, _, _, envs, count, _, _, files, user, debug, _ := getDeployParams(c)
	if entry == "" || image == "" {
		log.Fatalf("[Replace] no entry or image")
	}

	var data []byte
	if strings.HasPrefix(specURI, "http") {
		data, err = utils.GetSpecFromRemote(specURI)
	} else {
		data, err = ioutil.ReadFile(specURI)
	}
	if err != nil {
		return cli.Exit(err, -1)
	}

	force := c.Bool("force")
	labels := makeLabels(c.StringSlice("label"))
	deployOpts := generateDeployOpts(data, pod, node, entry, image, network, 0, 0, envs, count, nil, "", files, user, debug, false)
	return doReplaceContainer(client, deployOpts, force, labels)
}

func doReplaceContainer(client pb.CoreRPCClient, deployOpts *pb.DeployOptions, force bool, labels map[string]string) error {
	opts := &pb.ReplaceOptions{DeployOpt: deployOpts, Force: force, Labels: labels}
	resp, err := client.ReplaceContainer(context.Background(), opts)
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

		log.Infof("[Replace] Replace %s", msg.Remove.Id)
		if msg.Error != "" {
			log.Errorf("[Replace] Replace %s failed %s, message %s", msg.Remove.Id, msg.Error, msg.Remove.Message)
			if msg.Create != nil && msg.Create.Success {
				log.Errorf("[Replace] But create done id %s name %s", msg.Create.Id, msg.Create.Name)
			}
			continue
		} else if msg.Remove.Message != "" {
			log.Infof("[Replace] Other output \n%s", msg.Remove.Message)
		}

		// 一定会保证有 removeMsg 返回，success 一定为真
		removeMsg := msg.Remove
		log.Infof("[Replace] Old container %s removed", removeMsg.Id)

		// 到这里 create 肯定是成功了，否则错误会上浮到 err 中
		createMsg := msg.Create
		log.Infof("[Replace] New container %s, cpu %v, quota %v, mem %v", createMsg.Name, createMsg.Cpu, createMsg.Quota, createMsg.Memory)
		if len(createMsg.Hook) > 0 {
			log.Infof("[Replace] Other output \n%s", createMsg.Hook)
		}
		for name, publish := range createMsg.Publish {
			log.Infof("[Replace] Bound %s ip %s", name, publish)
		}
	}
	return nil
}
