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

type deployWorkloadsOptions struct {
	client      corepb.CoreRPCClient
	opts        *corepb.DeployOptions
	dryRun      bool
	autoReplace bool
}

func (o *deployWorkloadsOptions) run(ctx context.Context) error {
	if o.dryRun {
		r, err := o.client.CalculateCapacity(ctx, o.opts)
		if err != nil {
			return fmt.Errorf("[Deploy] Calculate capacity failed %v", err)
		}
		logrus.Infof("[Deploy] Capacity total %v", r.Total)
		for nodename, capacity := range r.NodeCapacities {
			logrus.Infof("[Deploy] Node %v capacity %v", nodename, capacity)
		}
		return nil
	}

	if !o.autoReplace {
		return doCreateContainer(ctx, o.client, o.opts)
	}

	lsOpts := &corepb.ListWorkloadsOptions{
		Appname:    o.opts.Name,
		Entrypoint: o.opts.Entrypoint.Name,
		Labels:     nil,
		Limit:      1, // 至少有一个可以被替换的
	}
	resp, err := o.client.ListWorkloads(ctx, lsOpts)
	if err != nil {
		return fmt.Errorf("[Deploy] check container failed %v", err)
	}
	_, err = resp.Recv()
	if err == io.EOF {
		logrus.Warn("[Deploy] there is no containers for replace")
		return doCreateContainer(ctx, o.client, o.opts)
	}
	if err != nil {
		return err
	}
	// 强制继承网络
	networkInherit := o.opts.Networkmode == ""
	return doReplaceContainer(ctx, o.client, o.opts, networkInherit, nil, nil)
}

func cmdWorkloadDeploy(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	for _, key := range []string{"pod", "entry", "image"} {
		if c.String(key) == "" {
			return fmt.Errorf("[Deploy] no %s given", key)
		}
	}
	if strings.Contains(c.String("entry"), "_") {
		return fmt.Errorf("[Deploy] entry can not contain _")
	}

	opts, err := utils.GenerateDeployOptions(c)
	if err != nil {
		return err
	}

	o := &deployWorkloadsOptions{
		client:      client,
		opts:        opts,
		dryRun:      c.Bool("dry-run"),
		autoReplace: c.Bool("auto-replace"),
	}
	return o.run(c.Context)
}

func doCreateContainer(ctx context.Context, client corepb.CoreRPCClient, deployOpts *corepb.DeployOptions) error {
	resp, err := client.CreateWorkload(ctx, deployOpts)
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

		if msg.Success {
			logrus.Infof("[Deploy] Success %s %s %s %v %v %v %d %d %v %v", msg.Id, msg.Name, msg.Nodename, msg.Resource.CpuQuotaRequest, msg.Resource.CpuQuotaLimit, msg.Resource.Cpu, msg.Resource.MemoryRequest, msg.Resource.MemoryLimit, msg.Resource.VolumePlanRequest, msg.Resource.VolumePlanLimit)
			if len(msg.Hook) > 0 {
				logrus.Infof("[Deploy] Hook output \n%s", msg.Hook)
			}
			for name, publish := range msg.Publish {
				logrus.Infof("[Deploy] Bound %s ip %s", name, publish)
			}
		} else {
			logrus.Errorf("[Deploy] Failed %v", msg.Error)
		}
	}
	return nil
}
