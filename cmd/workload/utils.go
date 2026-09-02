package workload

import (
	"context"
	"errors"
	"fmt"
	"strings"

	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/types"
)

func argIDs(cmd *cli.Command) ([]string, error) {
	ids := cmd.Args().Slice()
	if len(ids) == 0 {
		return nil, errors.New("workload id(s) should not be empty")
	}
	return ids, nil
}

func validateDeployFlags(cmd *cli.Command, keys ...string) error {
	for _, key := range keys {
		if cmd.String(key) == "" {
			return fmt.Errorf("no %s given", key)
		}
	}
	if strings.Contains(cmd.String(flagEntry), "_") {
		return errors.New("entry can not contain _")
	}
	return nil
}

func loadSpecs(ctx context.Context, cmd *cli.Command) (*types.Specs, error) {
	specURI := cmd.Args().First()
	if specURI == "" {
		return nil, errors.New("a spec must be given")
	}

	data, err := utils.ReadSpecURI(ctx, specURI)
	if err != nil {
		return nil, err
	}

	specs := &types.Specs{}
	if err := yaml.Unmarshal(data, specs); err != nil {
		return nil, fmt.Errorf("parse specs: %w", err)
	}
	return specs, nil
}

func entrypointOptions(specs *types.Specs, entry string) (*corepb.EntrypointOptions, error) {
	entrypoint, ok := specs.Entrypoints[entry]
	if !ok {
		return nil, fmt.Errorf("entry %s not found in specs", entry)
	}

	opts := &corepb.EntrypointOptions{
		Name:       entry,
		Commands:   entrypoint.GetCommands(),
		Privileged: entrypoint.Privileged,
		Dir:        entrypoint.Dir,
		Publish:    entrypoint.Publish,
		Restart:    entrypoint.Restart,
		Sysctls:    entrypoint.Sysctls,
	}
	if entrypoint.Hook != nil {
		opts.Hook = &corepb.HookOptions{
			AfterStart: entrypoint.Hook.AfterStart,
			BeforeStop: entrypoint.Hook.BeforeStop,
			Force:      entrypoint.Hook.Force,
		}
	}
	if entrypoint.HealthCheck != nil {
		opts.Healthcheck = &corepb.HealthCheckOptions{
			TcpPorts: entrypoint.HealthCheck.TCPPorts,
			HttpPort: entrypoint.HealthCheck.HTTPPort,
			Url:      entrypoint.HealthCheck.HTTPURL,
			Code:     int32(entrypoint.HealthCheck.HTTPCode), //nolint:gosec
		}
	}
	return opts, nil
}

func ramOption(cmd *cli.Command, request, limit, shortcut string) (int64, int64, error) {
	req, err := utils.ParseRAMInHuman(cmd.String(request))
	if err != nil {
		return 0, 0, err
	}
	lim, err := utils.ParseRAMInHuman(cmd.String(limit))
	if err != nil {
		return 0, 0, err
	}
	if cmd.IsSet(shortcut) {
		both, err := utils.ParseRAMInHuman(cmd.String(shortcut))
		return both, both, err
	}
	return req, lim, nil
}

func cpuOption(cmd *cli.Command) (float64, float64) {
	cpuRequest := cmd.Float64(flagCPURequest)
	cpuLimit := cmd.Float64(flagCPULimit)
	if cmd.IsSet("cpu") {
		cpu := cmd.Float64("cpu")
		cpuRequest, cpuLimit = cpu, cpu
	}
	return cpuRequest, cpuLimit
}

func cpumemParams(cmd *cli.Command, memoryRequest, memoryLimit int64) resourcetypes.RawParams {
	cpuRequest, cpuLimit := cpuOption(cmd)
	return resourcetypes.RawParams{
		flagCPURequest:    cpuRequest,
		flagCPULimit:      cpuLimit,
		flagMemoryRequest: memoryRequest,
		flagMemoryLimit:   memoryLimit,
	}
}
