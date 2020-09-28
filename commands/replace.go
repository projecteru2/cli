package commands

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/projecteru2/cli/utils"
	pb "github.com/projecteru2/core/rpc/gen"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	"golang.org/x/net/context"
)

func replaceContainers(c *cli.Context) error {
	client, err := checkParamsAndGetClient(c)
	if err != nil {
		return cli.Exit(err, -1)
	}
	specURI := c.Args().First()
	log.Debugf("[Replace] Replace container by %s", specURI)

	pod, entry, image, network, nodes, _, _, _, envs, count, _, _, files, user, debug, _, _, _, ignoreHook, afterCreate, _ := getDeployParams(c) // nolint
	if entry == "" || image == "" {
		log.Fatalf("[Replace] no entry or image")
	}
	copys := utils.SplitFiles(c.StringSlice("copy"))

	var data []byte
	if strings.HasPrefix(specURI, "http") {
		data, err = utils.GetSpecFromRemote(specURI)
	} else {
		data, err = ioutil.ReadFile(specURI)
	}
	if err != nil {
		return cli.Exit(err, -1)
	}

	labels := makeLabels(c.StringSlice("label"))
	networkInherit := c.Bool("network-inherit")
	// fix issue #15
	if network != "" {
		log.Warnf("[Replace] Network is not empty, so network-inherit will set to false")
		networkInherit = false
	}
	deployOpts := generateDeployOpts(data, pod, entry, image, network, nodes, 0, 0, 0, envs, count, nil, "", files, user, debug, false, false, ignoreHook, 0, afterCreate, "")
	return doReplaceContainer(client, deployOpts, networkInherit, labels, copys)
}

func doReplaceContainer(client pb.CoreRPCClient, deployOpts *pb.DeployOptions, networkInherit bool, labels map[string]string, copys map[string]string) error {
	opts := &pb.ReplaceOptions{DeployOpt: deployOpts, Networkinherit: networkInherit, FilterLabels: labels, Copy: copys}
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
			log.Errorf("[Replace] Replace %s failed %s, hook %s", msg.Remove.Id, msg.Error, msg.Remove.Hook)
			if msg.Create != nil && msg.Create.Success {
				log.Errorf("[Replace] But create done id %s name %s", msg.Create.Id, msg.Create.Name)
			}
			continue
		} else if msg.Remove.Hook != "" {
			log.Infof("[Replace] Hook output \n%s", msg.Remove.Hook)
		}

		// 一定会保证有 removeMsg 返回，success 一定为真
		removeMsg := msg.Remove
		log.Infof("[Replace] Hook container %s removed", removeMsg.Id)

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
