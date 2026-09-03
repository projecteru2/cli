package workload

import (
	"context"
	"errors"
	"strings"

	"github.com/projecteru2/core/log"
	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
)

type reallocWorkloadsOptions struct {
	client corepb.CoreRPCClient
	opts   *corepb.ReallocOptions
}

func (o *reallocWorkloadsOptions) run(ctx context.Context) error {
	resp, err := o.client.ReallocResource(ctx, o.opts)
	if err != nil {
		return err
	}

	if resp.Error != "" {
		return errors.New(resp.Error)
	}
	log.WithFunc("workload.reallocWorkloadsOptions.run").Info(ctx, "realloc success")
	return nil
}

func cmdWorkloadRealloc(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	opts, err := generateReallocOptions(cmd)
	if err != nil {
		return err
	}

	o := &reallocWorkloadsOptions{
		client: client,
		opts:   opts,
	}
	return o.run(ctx)
}

func generateReallocOptions(cmd *cli.Command) (*corepb.ReallocOptions, error) {
	id := cmd.Args().First()
	if id == "" {
		return nil, errors.New("workload id must be given")
	}

	memoryRequest, memoryLimit, err := ramOption(cmd, flagMemoryRequest, flagMemoryLimit, "memory")
	if err != nil {
		return nil, err
	}

	var volumesRequest, volumesLimit []string
	if v := cmd.String(flagVolumesRequest); v != "" {
		volumesRequest = strings.Split(v, ",")
	}
	if v := cmd.String(flagVolumesLimit); v != "" {
		volumesLimit = strings.Split(v, ",")
	}

	bindCPU := cmd.Bool("cpu-bind")
	unbindCPU := cmd.Bool("cpu-unbind")
	if bindCPU && unbindCPU {
		return nil, errors.New("cpu-bind and cpu-unbind can not both be set")
	}
	storageRequest, storageLimit, err := ramOption(cmd, flagStorageRequest, flagStorageLimit, flagStorage)
	if err != nil {
		return nil, err
	}

	cpumem := cpumemParams(cmd, memoryRequest, memoryLimit)
	switch {
	case bindCPU:
		cpumem["cpu-bind"] = true
	case !unbindCPU:
		cpumem["keep-cpu-bind"] = true
	}

	resources, err := utils.EncodeResources(cmd, resourcetypes.Resources{
		utils.ResourceCPUMem:  cpumem,
		utils.ResourceStorage: utils.StorageParams(storageRequest, storageLimit, volumesRequest, volumesLimit),
	})
	if err != nil {
		return nil, err
	}

	return &corepb.ReallocOptions{
		Id:        id,
		Resources: resources,
	}, nil
}
