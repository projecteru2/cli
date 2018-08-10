package commands

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/projecteru2/cli/utils"
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
	var data []byte
	if strings.HasPrefix(specURI, "http") {
		data, err = utils.GetSpecFromRemote(specURI)
	} else {
		data, err = ioutil.ReadFile(specURI)
	}
	if err != nil {
		return cli.Exit(err, -1)
	}

	opts := generateDeployOpts(data, pod, node, entry, image, network, 0, 0, envs, count, nil, "", files, user, debug, false)
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

		if msg.Error == "" {
			log.Infof("[Replace] Replace %s success", msg.Id)
			createMsg := msg.Create
			log.Infof("[Replace] New container %s, cpu %v, quota %v, mem %v", createMsg.Name, createMsg.Cpu, createMsg.Quota, createMsg.Memory)
			if len(createMsg.Hook) > 0 {
				log.Infof("[Replace] Hook output \n%s", createMsg.Hook)
			}
			for name, publish := range createMsg.Publish {
				log.Infof("[Replace] Bound %s ip %s", name, publish)
			}
		} else {
			log.Infof("[Replace] Replace %s failed %s", msg.Id, msg.Error)
		}
	}
	return nil
}
