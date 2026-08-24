package workload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	logger := log.WithFunc("workload.reallocWorkloadsOptions.run")
	if resp.Error != "" {
		logger.Error(ctx, errors.New(resp.Error), "realloc failed")
	} else {
		logger.Info(ctx, "realloc success")
	}
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
		return nil, errors.New("Workload ID must be given")
	}

	memoryRequest, memoryLimit, err := memoryOption(cmd)
	if err != nil {
		return nil, err
	}

	var volumesRequest, volumesLimit []string
	if v := cmd.String("volumes-request"); v != "" {
		volumesRequest = strings.Split(v, ",")
	}
	if v := cmd.String("volumes-limit"); v != "" {
		volumesLimit = strings.Split(v, ",")
	}

	bindCPU := cmd.Bool("cpu-bind")
	unbindCPU := cmd.Bool("cpu-unbind")
	if bindCPU && unbindCPU {
		return nil, errors.New("cpu-bind and cpu-unbind can not both be set")
	}
	bindCPUOpt := corepb.TriOpt_KEEP
	if bindCPU {
		bindCPUOpt = corepb.TriOpt_TRUE
	}
	if unbindCPU {
		bindCPUOpt = corepb.TriOpt_FALSE
	}

	storageRequest, storageLimit, err := storageOption(cmd)
	if err != nil {
		return nil, err
	}

	cpuRequest, cpuLimit := cpuOption(cmd)

	cpumem := resourcetypes.RawParams{
		"cpu-request":    cpuRequest,
		"cpu-limit":      cpuLimit,
		"memory-request": memoryRequest,
		"memory-limit":   memoryLimit,
	}
	storage := resourcetypes.RawParams{
		"storage-request": storageRequest,
		"storage-limit":   storageLimit,
		"volumes-request": volumesRequest,
		"volumes-limit":   volumesLimit,
	}

	switch bindCPUOpt {
	case corepb.TriOpt_KEEP:
		cpumem["keep-cpu-bind"] = true
	case corepb.TriOpt_TRUE:
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
		return nil, fmt.Errorf("[generateReallocOptions] get extra resources failed %v", err)
	}

	return &corepb.ReallocOptions{
		Id:        id,
		Resources: resources,
	}, nil
}
