package utils

import (
	"strings"

	"github.com/juju/errors"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
)

func GenerateReallocOptions(c *cli.Context) (*corepb.ReallocOptions, error) {
	id := c.Args().First()
	if id == "" {
		return nil, errors.New("Workload ID must be given")
	}

	memoryRequest, err := ParseRAMInHuman(c.String("memory-request"))
	if err != nil {
		return nil, err
	}
	memoryLimit, err := ParseRAMInHuman(c.String("memory-limit"))
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

	storageRequest, err := ParseRAMInHuman(c.String("storage-request"))
	if err != nil {
		return nil, err
	}
	storageLimit, err := ParseRAMInHuman(c.String("storage-limit"))
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
