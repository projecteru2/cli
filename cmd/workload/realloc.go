package workload

import (
	"context"
	"strings"

	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
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
		logrus.Infof("[Realloc] Failed by %+v", resp.Error)
	} else {
		logrus.Info("[Realloc] Success")
	}
	return nil
}

func cmdWorkloadRealloc(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	opts, err := generateReallocOptions(c)
	if err != nil {
		return err
	}

	o := &reallocWorkloadsOptions{
		client: client,
		opts:   opts,
	}
	return o.run(c.Context)
}

func generateReallocOptions(c *cli.Context) (*corepb.ReallocOptions, error) {
	id := c.Args().First()
	if id == "" {
		return nil, errors.New("Workload ID must be given")
	}

	memoryRequest, err := utils.ParseRAMInHuman(c.String("memory-request"))
	if err != nil {
		return nil, err
	}
	memoryLimit, err := utils.ParseRAMInHuman(c.String("memory-limit"))
	if err != nil {
		return nil, err
	}

	var volumesRequest, volumesLimit []string
	if v := c.String("volumes-request"); v != "" {
		volumesRequest = strings.Split(v, ",")
	}
	if v := c.String("volumes-limit"); v != "" {
		volumesLimit = strings.Split(v, ",")
	}

	bindCPU := c.Bool("cpu-bind")
	unbindCPU := c.Bool("cpu-unbind")
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

	storageRequest, err := utils.ParseRAMInHuman(c.String("storage-request"))
	if err != nil {
		return nil, err
	}
	storageLimit, err := utils.ParseRAMInHuman(c.String("storage-limit"))
	if err != nil {
		return nil, err
	}

	return &corepb.ReallocOptions{
		Id:         id,
		BindCpuOpt: bindCPUOpt,
		ResourceOpts: &corepb.ResourceOptions{
			CpuQuotaRequest: c.Float64("cpu-request"),
			CpuQuotaLimit:   c.Float64("cpu-limit"),
			MemoryRequest:   memoryRequest,
			MemoryLimit:     memoryLimit,
			VolumesRequest:  volumesRequest,
			VolumesLimit:    volumesLimit,
			StorageRequest:  storageRequest,
			StorageLimit:    storageLimit,
		},
	}, nil
}
