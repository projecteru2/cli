package workload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/projecteru2/core/log"
	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/types"
)

type deployWorkloadsOptions struct {
	client      corepb.CoreRPCClient
	opts        *corepb.DeployOptions
	dryRun      bool
	autoReplace bool
}

func (o *deployWorkloadsOptions) run(ctx context.Context) error {
	logger := log.WithFunc("workload.deployWorkloadsOptions.run")
	if o.dryRun {
		r, err := o.client.CalculateCapacity(ctx, o.opts)
		if err != nil {
			return fmt.Errorf("calculate capacity: %w", err)
		}
		logger.Infof(ctx, "capacity total %v", r.Total)
		for nodename, capacity := range r.NodeCapacities {
			logger.Infof(ctx, "node %v capacity %v", nodename, capacity)
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
		Limit:      1, // at least one workload must exist to be replaced
	}
	resp, err := o.client.ListWorkloads(ctx, lsOpts)
	if err != nil {
		return fmt.Errorf("check workload: %w", err)
	}
	_, err = resp.Recv()
	if errors.Is(err, io.EOF) {
		logger.Warn(ctx, "there is no workload to replace")
		return doCreateWorkload(ctx, o.client, o.opts)
	}
	if err != nil {
		return err
	}
	networkInherit := len(o.opts.Networks) == 0
	return doReplaceWorkload(ctx, o.client, o.opts, networkInherit, nil, nil)
}

func cmdWorkloadDeploy(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	for _, key := range []string{"pod", "entry", "image"} {
		if cmd.String(key) == "" {
			return fmt.Errorf("no %s given", key)
		}
	}
	if strings.Contains(cmd.String("entry"), "_") {
		return errors.New("entry can not contain _")
	}

	opts, err := generateDeployOptions(ctx, cmd)
	if err != nil {
		return err
	}

	o := &deployWorkloadsOptions{
		client:      client,
		opts:        opts,
		dryRun:      cmd.Bool("dry-run"),
		autoReplace: cmd.Bool("auto-replace"),
	}
	return o.run(ctx)
}

func doCreateWorkload(ctx context.Context, client corepb.CoreRPCClient, deployOpts *corepb.DeployOptions) error {
	logger := log.WithFunc("workload.doCreateWorkload")
	resp, err := client.CreateWorkload(ctx, deployOpts)
	if err != nil {
		return err
	}
	for {
		msg, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		if msg.Success {
			logger.Infof(ctx, "create %s %s on %s success, resource: %s", msg.Id, msg.Name, msg.Nodename, msg.Resources)
			if len(msg.Hook) > 0 {
				logger.Infof(ctx, "hook output \n%s", msg.Hook)
			}
			for name, publish := range msg.Publish {
				logger.Infof(ctx, "bound %s ip %s", name, publish)
			}
		} else {
			logger.Error(ctx, errors.New(msg.Error), "create workload failed")
		}
	}
	return nil
}

func generateDeployOptions(ctx context.Context, cmd *cli.Command) (*corepb.DeployOptions, error) {
	specURI := cmd.Args().First()
	if specURI == "" {
		return nil, errors.New("a spec must be given")
	}
	log.WithFunc("workload.generateDeployOptions").Debugf(ctx, "deploy %s", specURI)

	var (
		data []byte
		err  error
	)
	if strings.HasPrefix(specURI, "http") {
		data, err = utils.GetSpecFromRemote(ctx, specURI)
	} else {
		data, err = os.ReadFile(specURI) //nolint:gosec
	}
	if err != nil {
		return nil, err
	}

	memoryRequest, memoryLimit, err := memoryOption(cmd)
	if err != nil {
		return nil, fmt.Errorf("parse memory: %w", err)
	}

	storageRequest, storageLimit, err := storageOption(cmd)
	if err != nil {
		return nil, fmt.Errorf("parse storage: %w", err)
	}

	cpuRequest, cpuLimit := cpuOption(cmd)

	specs := &types.Specs{}
	if err := yaml.Unmarshal(data, specs); err != nil {
		return nil, fmt.Errorf("parse specs: %w", err)
	}

	entry := cmd.String("entry")

	network := cmd.String("network")
	networks := utils.GetNetworks(network)
	entrypoint, ok := specs.Entrypoints[entry]
	if !ok {
		return nil, fmt.Errorf("entry %s not found in specs", entry)
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
			Code:     int32(entrypoint.HealthCheck.HTTPCode), //nolint:gosec
		}
	}

	var logConfig *corepb.LogOptions
	if entrypoint.Log != nil {
		logConfig = &corepb.LogOptions{
			Type:   entrypoint.Log.Type,
			Config: entrypoint.Log.Config,
		}
	}

	rawArgs := cmd.String("raw-args")
	rawArgsByte := []byte{}
	if rawArgs != "" {
		rawArgsByte = []byte(rawArgs)
	}

	content, modes, owners := utils.GenerateFileOptions(cmd)

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

	if cmd.Bool("cpu-bind") {
		cpumem["cpu-bind"] = true
	}

	cb, _ := json.Marshal(cpumem)
	sb, _ := json.Marshal(storage)

	resources := map[string][]byte{
		"cpumem":  cb,
		"storage": sb,
	}

	if extraResourcesMap, err := utils.ParseExtraResources(cmd); err == nil {
		for k, v := range extraResourcesMap {
			if _, ok := resources[k]; ok {
				continue
			}
			eb, _ := json.Marshal(v)
			resources[k] = eb
		}
	} else {
		return nil, fmt.Errorf("parse extra resources: %w", err)
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
		Podname:   cmd.String("pod"),
		NodeFilter: &corepb.NodeFilter{
			Includes: cmd.StringSlice("node"),
			Labels:   utils.SplitEquality(cmd.StringSlice("nodelabel")),
		},
		Image:          cmd.String("image"),
		Count:          int32(cmd.Int("count")), //nolint:gosec
		Env:            cmd.StringSlice("env"),
		Networks:       networks,
		Labels:         specs.Labels,
		Dns:            specs.DNS,
		ExtraHosts:     specs.ExtraHosts,
		DeployStrategy: corepb.DeployOptions_Strategy(corepb.DeployOptions_Strategy_value[strings.ToUpper(cmd.String("deploy-strategy"))]),
		Data:           content,
		Modes:          modes,
		Owners:         owners,
		User:           cmd.String("user"),
		Debug:          cmd.Bool("debug"),
		NodesLimit:     int32(cmd.Int("nodes-limit")), //nolint:gosec
		IgnoreHook:     cmd.Bool("ignore-hook"),
		AfterCreate:    cmd.StringSlice("after-create"),
		RawArgs:        rawArgsByte,
	}, nil
}
