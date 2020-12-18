package workload

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

type replaceWorkloadsOptions struct {
	client         corepb.CoreRPCClient
	opts           *corepb.DeployOptions
	labels         map[string]string
	copys          map[string]string
	networkInherit bool
}

func (o *replaceWorkloadsOptions) run(ctx context.Context) error {
	return doReplaceWorkload(ctx, o.client, o.opts, o.networkInherit, o.labels, o.copys)
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

	opts, err := generateReplaceOptions(c)
	if err != nil {
		return err
	}

	networkInherit := c.Bool("network-inherit")
	if len(opts.Networks) > 0 {
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

func doReplaceWorkload(ctx context.Context, client corepb.CoreRPCClient, deployOpts *corepb.DeployOptions, networkInherit bool, labels map[string]string, copys map[string]string) error {
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
		logrus.Infof("[Replace] Hook workload %s removed", removeMsg.Id)

		// 到这里 create 肯定是成功了，否则错误会上浮到 err 中
		createMsg := msg.Create
		logrus.Infof("[Replace] New workload %s, cpu %v, quotaRequest %v, quotaLimit %v, memRequest %v, memLimit %v", createMsg.Name, createMsg.Resource.Cpu, createMsg.Resource.CpuQuotaRequest, createMsg.Resource.CpuQuotaLimit, createMsg.Resource.MemoryRequest, createMsg.Resource.MemoryLimit)
		if len(createMsg.Hook) > 0 {
			logrus.Infof("[Replace] Other output \n%s", createMsg.Hook)
		}
		for name, publish := range createMsg.Publish {
			logrus.Infof("[Replace] Bound %s ip %s", name, publish)
		}
	}
	return nil
}

func generateReplaceOptions(c *cli.Context) (*corepb.DeployOptions, error) {
	specURI := c.Args().First()
	if specURI == "" {
		return nil, fmt.Errorf("a specs must be given")
	}
	logrus.Debugf("[Replace] Replace with %s", specURI)

	var (
		data []byte
		err  error
	)
	if strings.HasPrefix(specURI, "http") {
		data, err = utils.GetSpecFromRemote(specURI)
	} else {
		data, err = ioutil.ReadFile(specURI)
	}
	if err != nil {
		return nil, err
	}

	specs := &types.Specs{}
	if err := yaml.Unmarshal(data, specs); err != nil {
		return nil, fmt.Errorf("[generateReplaceOptions] get specs failed %v", err)
	}

	entry := c.String("entry")

	network := c.String("network")
	networks := utils.GetNetworks(network)
	entrypoint, ok := specs.Entrypoints[entry]
	if !ok {
		return nil, fmt.Errorf("[generateReplaceOptions] get entry failed")
	}

	var hook *corepb.HookOptions
	if entrypoint.Hook != nil {
		hook = &corepb.HookOptions{
			AfterStart: entrypoint.Hook.AfterStart,
			BeforeStop: entrypoint.Hook.BeforeStop,
			Force:      entrypoint.Hook.Force,
		}
	}

	var healthCheck *corepb.HealthCheckOptions
	if entrypoint.HealthCheck != nil {
		healthCheck = &corepb.HealthCheckOptions{
			TcpPorts: entrypoint.HealthCheck.TCPPorts,
			HttpPort: entrypoint.HealthCheck.HTTPPort,
			Url:      entrypoint.HealthCheck.HTTPURL,
			Code:     int32(entrypoint.HealthCheck.HTTPCode),
		}
	}

	var logConfig *corepb.LogOptions
	if entrypoint.Log != nil {
		logConfig = &corepb.LogOptions{
			Type:   entrypoint.Log.Type,
			Config: entrypoint.Log.Config,
		}
	}

	return &corepb.DeployOptions{
		Name: specs.Appname,
		Entrypoint: &corepb.EntrypointOptions{
			Name:          entry,
			Command:       entrypoint.Command,
			Privileged:    entrypoint.Privileged,
			Dir:           entrypoint.Dir,
			Log:           logConfig,
			Publish:       entrypoint.Publish,
			Healthcheck:   healthCheck,
			Hook:          hook,
			RestartPolicy: entrypoint.RestartPolicy,
			Sysctls:       entrypoint.Sysctls,
		},
		ResourceOpts: &corepb.ResourceOptions{
			CpuQuotaRequest: 0,
			CpuQuotaLimit:   0,
			CpuBind:         false,
			MemoryRequest:   0,
			MemoryLimit:     0,
			StorageRequest:  0,
			StorageLimit:    0,
			VolumesRequest:  specs.VolumesRequest,
			VolumesLimit:    specs.Volumes,
		},
		Podname:        c.String("pod"),
		Nodenames:      c.StringSlice("node"),
		Image:          c.String("image"),
		Count:          int32(c.Int("count")),
		Env:            c.StringSlice("env"),
		Networks:       networks,
		Labels:         specs.Labels,
		Dns:            specs.DNS,
		ExtraHosts:     specs.ExtraHosts,
		Nodelabels:     nil,
		DeployStrategy: corepb.DeployOptions_Strategy(corepb.DeployOptions_Strategy_value[""]),
		Data:           utils.ReadAllFiles(c.StringSlice("file")),
		User:           c.String("user"),
		Debug:          c.Bool("debug"),
		NodesLimit:     0,
		IgnoreHook:     c.Bool("ignore-hook"),
		AfterCreate:    c.StringSlice("after-create"),
		RawArgs:        []byte{},
	}, nil
}
