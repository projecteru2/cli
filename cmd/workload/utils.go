package workload

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/types"
)

func loadSpecs(ctx context.Context, cmd *cli.Command) (*types.Specs, error) {
	specURI := cmd.Args().First()
	if specURI == "" {
		return nil, errors.New("a spec must be given")
	}

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
	if entrypoint.Log != nil {
		opts.Log = &corepb.LogOptions{
			Type:   entrypoint.Log.Type,
			Config: entrypoint.Log.Config,
		}
	}
	return opts, nil
}

func memoryOption(cmd *cli.Command) (int64, int64, error) {
	memRequest, err := utils.ParseRAMInHuman(cmd.String(flagMemoryRequest))
	if err != nil {
		return 0, 0, err
	}

	memLimit, err := utils.ParseRAMInHuman(cmd.String(flagMemoryLimit))
	if err != nil {
		return 0, 0, err
	}
	if cmd.IsSet("memory") {
		memory, err := utils.ParseRAMInHuman(cmd.String("memory"))
		if err != nil {
			return 0, 0, err
		}
		memRequest, memLimit = memory, memory
	}
	return memRequest, memLimit, nil
}

func storageOption(cmd *cli.Command) (int64, int64, error) {
	storageRequest, err := utils.ParseRAMInHuman(cmd.String(flagStorageRequest))
	if err != nil {
		return 0, 0, err
	}

	storageLimit, err := utils.ParseRAMInHuman(cmd.String(flagStorageLimit))
	if err != nil {
		return 0, 0, err
	}
	if cmd.IsSet(flagStorage) {
		storage, err := utils.ParseRAMInHuman(cmd.String(flagStorage))
		if err != nil {
			return 0, 0, err
		}
		storageRequest, storageLimit = storage, storage
	}
	return storageRequest, storageLimit, nil
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
