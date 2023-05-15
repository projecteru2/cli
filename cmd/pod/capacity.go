package pod

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/projecteru2/cli/cmd/utils"
	"github.com/projecteru2/cli/describe"
	resourcetypes "github.com/projecteru2/core/resource/types"
	corepb "github.com/projecteru2/core/rpc/gen"

	"github.com/google/uuid"
	"github.com/juju/errors"
	"github.com/urfave/cli/v2"
)

type capacityPodOptions struct {
	client    corepb.CoreRPCClient
	podname   string   // podname
	nodenames []string // node white list

	cpu            float64
	cpuBind        bool
	memory         int64
	storage        int64
	extraResources map[string]any
}

func (o *capacityPodOptions) run(ctx context.Context) error {
	cpumem := resourcetypes.RawParams{
		"cpu":    o.cpu,
		"memory": o.memory,
	}
	storage := resourcetypes.RawParams{
		"storage": o.storage,
	}

	if o.cpuBind {
		cpumem["cpu-bind"] = true
	}

	cb, _ := json.Marshal(cpumem)
	sb, _ := json.Marshal(storage)
	resources := map[string][]byte{
		"cpumem":  cb,
		"storage": sb,
	}

	for k, v := range o.extraResources {
		if _, ok := resources[k]; ok {
			continue
		}
		eb, _ := json.Marshal(v)
		resources[k] = eb
	}

	opts := &corepb.DeployOptions{
		// resource definitions
		Resources: resources,

		// deploy options
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

	extraResourcesMap, err := utils.ParseExtraResources(c)
	if err != nil {
		return fmt.Errorf("[cmdPodCapacity] parse extra resources failed %v", err)
	}

	o := &capacityPodOptions{
		client:    client,
		podname:   name,
		nodenames: c.StringSlice("node"),

		cpu:            c.Float64("cpu"),
		cpuBind:        c.Bool("cpu-bind"),
		memory:         mem,
		storage:        storage,
		extraResources: extraResourcesMap,
	}
	return o.run(c.Context)
}
