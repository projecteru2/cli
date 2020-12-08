package workload

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/projecteru2/cli/pkg/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type replaceWorkloadsOptions struct {
	client         corepb.CoreRPCClient
	opts           *corepb.DeployOptions
	labels         map[string]string
	copys          map[string]string
	networkInherit bool
}

func (o *replaceWorkloadsOptions) run(ctx context.Context) error {
	return doReplaceContainer(ctx, o.client, o.opts, o.networkInherit, o.labels, o.copys)
}

func cmdWorkloadReplace(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	for _, key := range []string{"entry", "image"} {
		if c.String(key) == "" {
			return fmt.Errorf("[Replace] no %s given", key)
		}
	}
	if strings.Contains(c.String("entry"), "_") {
		return fmt.Errorf("[Replace] entry can not contain _")
	}

	opts, err := utils.GenerateReplaceOptions(c)
	if err != nil {
		return err
	}

	networkInherit := c.Bool("network-inherit")
	if opts.Networkmode != "" {
		logrus.Warnf("[Replace] Network is not empty, so network-inherit will set to false")
		networkInherit = false
	}
	o := &replaceWorkloadsOptions{
		client:         client,
		opts:           opts,
		copys:          utils.SplitFiles(c.StringSlice("copy")),
		labels:         utils.SplitEquality(c.StringSlice("label")),
		networkInherit: networkInherit,
	}
	return o.run(c.Context)
}

func doReplaceContainer(ctx context.Context, client corepb.CoreRPCClient, deployOpts *corepb.DeployOptions, networkInherit bool, labels map[string]string, copys map[string]string) error {
	opts := &corepb.ReplaceOptions{
		DeployOpt:      deployOpts,
		Networkinherit: networkInherit,
		FilterLabels:   labels,
		Copy:           copys,
	}
	resp, err := client.ReplaceWorkload(ctx, opts)
	if err != nil {
		return err
	}
	for {
		msg, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		logrus.Infof("[Replace] Replace %s", msg.Remove.Id)
		if msg.Error != "" {
			logrus.Errorf("[Replace] Replace %s failed %s, hook %s", msg.Remove.Id, msg.Error, msg.Remove.Hook)
			if msg.Create != nil && msg.Create.Success {
				logrus.Errorf("[Replace] But create done id %s name %s", msg.Create.Id, msg.Create.Name)
			}
			continue
		} else if msg.Remove.Hook != "" {
			logrus.Infof("[Replace] Hook output \n%s", msg.Remove.Hook)
		}

		// 一定会保证有 removeMsg 返回，success 一定为真
		removeMsg := msg.Remove
		logrus.Infof("[Replace] Hook container %s removed", removeMsg.Id)

		// 到这里 create 肯定是成功了，否则错误会上浮到 err 中
		createMsg := msg.Create
		logrus.Infof("[Replace] New container %s, cpu %v, quotaRequest %v, quotaLimit %v, memRequest %v, memLimit %v", createMsg.Name, createMsg.Resource.Cpu, createMsg.Resource.CpuQuotaRequest, createMsg.Resource.CpuQuotaLimit, createMsg.Resource.MemoryRequest, createMsg.Resource.MemoryLimit)
		if len(createMsg.Hook) > 0 {
			logrus.Infof("[Replace] Other output \n%s", createMsg.Hook)
		}
		for name, publish := range createMsg.Publish {
			logrus.Infof("[Replace] Bound %s ip %s", name, publish)
		}
	}
	return nil
}
