package workload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/projecteru2/core/log"
	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
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
		Limit:      1,
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

	for _, key := range []string{flagPod, flagEntry, flagImage} {
		if cmd.String(key) == "" {
			return fmt.Errorf("no %s given", key)
		}
	}
	if strings.Contains(cmd.String(flagEntry), "_") {
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
	specs, err := loadSpecs(ctx, cmd)
	if err != nil {
		return nil, err
	}

	entrypoint, err := entrypointOptions(specs, cmd.String(flagEntry))
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

	cpumem := resourcetypes.RawParams{
		flagCPURequest:    cpuRequest,
		flagCPULimit:      cpuLimit,
		flagMemoryRequest: memoryRequest,
		flagMemoryLimit:   memoryLimit,
	}
	if cmd.Bool("cpu-bind") {
		cpumem["cpu-bind"] = true
	}
	storage := resourcetypes.RawParams{
		flagStorageRequest: storageRequest,
		flagStorageLimit:   storageLimit,
		flagVolumesRequest: specs.VolumesRequest,
		flagVolumesLimit:   specs.Volumes,
	}

	resources, err := utils.EncodeResources(cmd, resourcetypes.Resources{
		resourceCPUMem:  cpumem,
		resourceStorage: storage,
	})
	if err != nil {
		return nil, err
	}

	files, err := utils.GenerateFileOptions(cmd)
	if err != nil {
		return nil, err
	}

	deployStrategy, err := utils.ParseDeployStrategy(cmd.String("deploy-strategy"))
	if err != nil {
		return nil, err
	}

	return &corepb.DeployOptions{
		Name:       specs.Appname,
		Entrypoint: entrypoint,
		Resources:  resources,
		Podname:    cmd.String(flagPod),
		NodeFilter: &corepb.NodeFilter{
			Includes: cmd.StringSlice(flagNode),
			Labels:   utils.SplitEquality(cmd.StringSlice("nodelabel")),
		},
		Image:          cmd.String(flagImage),
		Count:          int32(cmd.Int("count")), //nolint:gosec
		Env:            cmd.StringSlice(flagEnv),
		Networks:       utils.GetNetworks(cmd.String(flagNetwork)),
		Labels:         specs.Labels,
		Dns:            specs.DNS,
		ExtraHosts:     specs.ExtraHosts,
		DeployStrategy: deployStrategy,
		Data:           files.Data,
		Modes:          files.Modes,
		Owners:         files.Owners,
		User:           cmd.String("user"),
		Debug:          cmd.Bool("debug"),
		NodesLimit:     int32(cmd.Int("nodes-limit")), //nolint:gosec
		IgnoreHook:     cmd.Bool("ignore-hook"),
		AfterCreate:    cmd.StringSlice("after-create"),
		RawArgs:        []byte(cmd.String("raw-args")),
	}, nil
}
