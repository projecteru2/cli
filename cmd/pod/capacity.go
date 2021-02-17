package pod

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/juju/errors"
	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v2"
)

type capacityPodOptions struct {
	client corepb.CoreRPCClient
	name   string

	cpu     float64
	cpuBind bool
	memory  int64
	storage int64
}

func (o *capacityPodOptions) run(ctx context.Context) error {
	resp, err := o.client.CalculateCapacity(ctx, &corepb.DeployOptions{
		// resource definitions
		ResourceOpts: &corepb.ResourceOptions{
			CpuQuotaLimit:   o.cpu,
			CpuQuotaRequest: o.cpu,
			CpuBind:         o.cpuBind,
			MemoryLimit:     o.memory,
			MemoryRequest:   o.memory,
			StorageLimit:    o.storage,
			StorageRequest:  o.storage,
		},

		// deploy options
		Entrypoint: &corepb.EntrypointOptions{
			Name: uuid.New().String(),
		},
		DeployStrategy: corepb.DeployOptions_DUMMY,
		Podname:        o.name,
	})
	if err != nil {
		return err
	}

	describe.PodCapacity(resp.Total, resp.NodeCapacities)
	return nil
}

func cmdPodCapacity(c *cli.Context) error {
	client, err := utils.NewCoreRPCClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if name == "" {
		return errors.New("Pod name must be given")
	}

	mem, err := utils.ParseRAMInHuman(c.String("memory"))
	if err != nil {
		return fmt.Errorf("[cmdPodCapacity] parse memory failed %v", err)
	}

	storage, err := utils.ParseRAMInHuman(c.String("storage"))
	if err != nil {
		return fmt.Errorf("[cmdPodCapacity] parse storage failed %v", err)
	}

	o := &capacityPodOptions{
		client: client,
		name:   name,

		cpu:     c.Float64("cpu-limit"),
		cpuBind: c.Bool("cpu-bind"),
		memory:  mem,
		storage: storage,
	}
	return o.run(c.Context)
}
