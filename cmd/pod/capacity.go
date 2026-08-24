package pod

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"
	"github.com/urfave/cli/v3"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
)

type capacityPodOptions struct {
	client    corepb.CoreRPCClient
	podname   string
	nodenames []string // node white list

	cpu            float64
	cpuBind        bool
	memory         int64
	storage        int64
	extraResources map[string]any
}

func (o *capacityPodOptions) run(ctx context.Context) error {
	cpumem := resourcetypes.RawParams{
		flagCPU:    o.cpu,
		flagMemory: o.memory,
	}
	storage := resourcetypes.RawParams{
		flagStorage: o.storage,
	}

	if o.cpuBind {
		cpumem["cpu-bind"] = true
	}

	cb, _ := json.Marshal(cpumem)
	sb, _ := json.Marshal(storage)
	resources := map[string][]byte{
		resourceCPUMem:  cb,
		resourceStorage: sb,
	}

	for k, v := range o.extraResources {
		if _, ok := resources[k]; ok {
			continue
		}
		eb, _ := json.Marshal(v)
		resources[k] = eb
	}

	opts := &corepb.DeployOptions{
		Resources: resources,
		Entrypoint: &corepb.EntrypointOptions{
			Name: uuid.New().String(),
		},
		DeployStrategy: corepb.DeployOptions_DUMMY,
		Podname:        o.podname,
		NodeFilter: &corepb.NodeFilter{
			Includes: o.nodenames,
		},
	}

	resp, err := o.client.CalculateCapacity(ctx, opts)
	if err != nil {
		return err
	}

	describe.PodCapacity(resp.Total, resp.NodeCapacities)
	return nil
}

func cmdPodCapacity(ctx context.Context, cmd *cli.Command) error {
	client, err := utils.NewCoreRPCClient(ctx, cmd)
	if err != nil {
		return err
	}

	name := cmd.Args().First()
	if name == "" {
		return errors.New("pod name must be given")
	}

	mem, err := utils.ParseRAMInHuman(cmd.String(flagMemory))
	if err != nil {
		return fmt.Errorf("parse memory: %w", err)
	}

	storage, err := utils.ParseRAMInHuman(cmd.String(flagStorage))
	if err != nil {
		return fmt.Errorf("parse storage: %w", err)
	}

	extraResourcesMap, err := utils.ParseExtraResources(cmd)
	if err != nil {
		return fmt.Errorf("parse extra resources: %w", err)
	}

	o := &capacityPodOptions{
		client:    client,
		podname:   name,
		nodenames: cmd.StringSlice("node"),

		cpu:            cmd.Float64(flagCPU),
		cpuBind:        cmd.Bool("cpu-bind"),
		memory:         mem,
		storage:        storage,
		extraResources: extraResourcesMap,
	}
	return o.run(ctx)
}
