package workload

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/types"
	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
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
		return doCreateWorkload(ctx, o.client, o.opts)
	}

	lsOpts := &corepb.ListWorkloadsOptions{
		Appname:    o.opts.Name,
		Entrypoint: o.opts.Entrypoint.Name,
		Labels:     nil,
		Limit:      1, // 至少有一个可以被替换的
	}
	resp, err := o.client.ListWorkloads(ctx, lsOpts)
	if err != nil {
		return fmt.Errorf("[Deploy] check workload failed %v", err)
	}
	_, err = resp.Recv()
	if err == io.EOF {
		logrus.Warn("[Deploy] there is no Workloads for replace")
		return doCreateWorkload(ctx, o.client, o.opts)
	}
	if err != nil {
		return err
	}
	// 强制继承网络
	networkInherit := len(o.opts.Networks) == 0
	return doReplaceWorkload(ctx, o.client, o.opts, networkInherit, nil, nil)
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

	opts, err := generateDeployOptions(c)
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

func doCreateWorkload(ctx context.Context, client corepb.CoreRPCClient, deployOpts *corepb.DeployOptions) error {
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
			logrus.Infof("[Deploy] Success %s %s %s %s", msg.Id, msg.Name, msg.Nodename, msg.Resources)
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

func generateDeployOptions(c *cli.Context) (*corepb.DeployOptions, error) {
	specURI := c.Args().First()
	if specURI == "" {
		return nil, fmt.Errorf("a specs must be given")
	}
	logrus.Debugf("[Deploy] Deploy %s", specURI)

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

	memoryRequest, memoryLimit, err := memoryOption(c)
	if err != nil {
		return nil, fmt.Errorf("[generateDeployOptions] parse memory failed %v", err)
	}

	storageRequest, storageLimit, err := storageOption(c)
	if err != nil {
		return nil, fmt.Errorf("[generateDeployOptions] parse storage failed %v", err)
	}

	cpuRequest, cpuLimit := cpuOption(c)

	specs := &types.Specs{}
	if err := yaml.Unmarshal(data, specs); err != nil {
		return nil, fmt.Errorf("[generateDeployOptions] get specs failed %v", err)
	}

	entry := c.String("entry")

	network := c.String("network")
	networks := utils.GetNetworks(network)
	entrypoint, ok := specs.Entrypoints[entry]
	if !ok {
		return nil, fmt.Errorf("[generateDeployOptions] get entry failed")
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

	rawArgs := c.String("raw-args")
	rawArgsByte := []byte{}
	if rawArgs != "" {
		rawArgsByte = []byte(rawArgs)
	}

	content, modes, owners := utils.GenerateFileOptions(c)

	cpumem := resourcetypes.RawParams{
		"cpu-request":    cpuRequest,
		"cpu-limit":      cpuLimit,
		"memory-request": memoryRequest,
		"memory-limit":   memoryLimit,
	}
	storage := resourcetypes.RawParams{
		"storage-request": storageRequest,
		"storage-limit":   storageLimit,
		"volumes-request": specs.VolumesRequest,
		"volumes-limit":   specs.Volumes,
	}

	if c.Bool("cpu-bind") {
		cpumem["cpu-bind"] = true
	}

	cb, _ := json.Marshal(cpumem)
	sb, _ := json.Marshal(storage)

	resources := map[string][]byte{
		"cpumem":  cb,
		"storage": sb,
	}

	extraResources := c.String("extra-resources")
	if extraResources != "" {
		extraResourcesMap := make(map[string]any)
		if err := json.Unmarshal([]byte(extraResources), &extraResourcesMap); err != nil {
			return nil, fmt.Errorf("Invalid value for extra-resources: %v", err)
		}
		for k, v := range extraResourcesMap {
			if _, ok := resources[k]; ok {
				continue
			}
			eb, _ := json.Marshal(v)
			resources[k] = eb
		}
	}
	return &corepb.DeployOptions{
		Name: specs.Appname,
		Entrypoint: &corepb.EntrypointOptions{
			Name:        entry,
			Commands:    entrypoint.GetCommands(),
			Privileged:  entrypoint.Privileged,
			Dir:         entrypoint.Dir,
			Log:         logConfig,
			Publish:     entrypoint.Publish,
			Healthcheck: healthCheck,
			Hook:        hook,
			Restart:     entrypoint.Restart,
			Sysctls:     entrypoint.Sysctls,
		},
		Resources: resources,
		Podname:   c.String("pod"),
		NodeFilter: &corepb.NodeFilter{
			Includes: c.StringSlice("node"),
			Labels:   utils.SplitEquality(c.StringSlice("nodelabel")),
		},
		Image:          c.String("image"),
		Count:          int32(c.Int("count")),
		Env:            c.StringSlice("env"),
		Networks:       networks,
		Labels:         specs.Labels,
		Dns:            specs.DNS,
		ExtraHosts:     specs.ExtraHosts,
		DeployStrategy: corepb.DeployOptions_Strategy(corepb.DeployOptions_Strategy_value[strings.ToUpper(c.String("deploy-strategy"))]),
		Data:           content,
		Modes:          modes,
		Owners:         owners,
		User:           c.String("user"),
		Debug:          c.Bool("debug"),
		NodesLimit:     int32(c.Int("nodes-limit")),
		IgnoreHook:     c.Bool("ignore-hook"),
		AfterCreate:    c.StringSlice("after-create"),
		RawArgs:        rawArgsByte,
	}, nil
}
