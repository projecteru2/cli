package workload

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/projecteru2/core/log"
	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type deployWorkloadsOptions struct {
	client         corepb.CoreRPCClient
	opts           *corepb.DeployOptions
	dryRun         bool
	autoReplace    bool
	networkInherit bool
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
	}
	probeCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	resp, err := o.client.ListWorkloads(probeCtx, lsOpts)
	if err != nil {
		return fmt.Errorf("check workload: %w", err)
	}
	for {
		workload, recvErr := resp.Recv()
		if errors.Is(recvErr, io.EOF) {
			logger.Warn(ctx, "there is no workload in the target pod to replace")
			return doCreateWorkload(ctx, o.client, o.opts)
		}
		if recvErr != nil {
			return recvErr
		}
		if workload.Podname == o.opts.Podname {
			cancel()
			return doReplaceWorkload(ctx, o.client, o.opts, o.networkInherit, nil, nil)
		}
	}
}

func cmdWorkloadDeploy(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	if err = validateDeployFlags(cmd, flagPod, flagEntry, flagImage); err != nil {
		return err
	}

	opts, err := generateDeployOptions(ctx, cmd)
	if err != nil {
		return err
	}

	o := &deployWorkloadsOptions{
		client:         client,
		opts:           opts,
		dryRun:         cmd.Bool("dry-run"),
		autoReplace:    cmd.Bool("auto-replace"),
		networkInherit: !cmd.IsSet(flagNetwork),
	}
	return o.run(ctx)
}

func doCreateWorkload(ctx context.Context, client corepb.CoreRPCClient, deployOpts *corepb.DeployOptions) error {
	logger := log.WithFunc("workload.doCreateWorkload")
	resp, err := client.CreateWorkload(ctx, deployOpts)
	if err != nil {
		return err
	}
	return utils.EachMessage(resp.Recv, func(msg *corepb.CreateWorkloadMessage) error {
		if !msg.Success {
			return fmt.Errorf("create workload on %s: %s", msg.Nodename, msg.Error)
		}

		logger.Infof(ctx, "create %s %s on %s success, resource: %s", msg.Id, msg.Name, msg.Nodename, msg.Resources)
		if len(msg.Hook) > 0 {
			logger.Infof(ctx, "hook output \n%s", msg.Hook)
		}
		for name, publish := range msg.Publish {
			logger.Infof(ctx, "bound %s ip %s", name, publish)
		}
		return nil
	})
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

	memoryRequest, memoryLimit, err := ramOption(cmd, flagMemoryRequest, flagMemoryLimit, "memory")
	if err != nil {
		return nil, fmt.Errorf("parse memory: %w", err)
	}

	storageRequest, storageLimit, err := ramOption(cmd, flagStorageRequest, flagStorageLimit, flagStorage)
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
		utils.ResourceCPUMem:  cpumem,
		utils.ResourceStorage: storage,
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
